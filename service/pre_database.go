package main

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"runtime"

	"github.com/gin-gonic/gin"
	jsonIteratorExtra "github.com/json-iterator/go/extra"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
	"github.com/v2rayA/v2rayA/kernel/v2ray/asset"
	"github.com/v2rayA/v2rayA/kernel/v2ray/asset/dat"
	"github.com/v2rayA/v2rayA/kernel/v2ray/where"
	"github.com/v2rayA/v2rayA/pkg/util/copyfile"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"github.com/v2rayA/v2rayA/server/service"

	confv4 "github.com/v2rayA/v2rayA-lib4/conf"
	touchv4 "github.com/v2rayA/v2rayA-lib4/core/touch"
	configurev4 "github.com/v2rayA/v2rayA-lib4/db/configure"
	servicev4 "github.com/v2rayA/v2rayA-lib4/server/service"
)

func initDBValue() {
	log.Info("init DB")
	err := configure.SetConfigure(configure.New())
	if err != nil {
		log.Fatal("initDBValue: %v", err)
	}
}

func initConfigure() {
	// initialize configuration
	jsonIteratorExtra.RegisterFuzzyDecoders()

	// Track whether we performed a BoltDB→SQLite migration in this session.
	// If so, we must skip the v4 migration and initDBValue() below, because
	// the data has already been migrated into SQLite by MigrateFromBoltDB().
	migratedFromBoltDB := false

	// Try to initialize SQLite database.
	// If an old BoltDB database (bolt.db) exists, Open() returns ErrNeedMigration,
	// and we must run the migration first.
	err := db.Open()
	if errors.Is(err, db.ErrNeedMigration) {
		log.Warn("Detected legacy BoltDB database, migrating to SQLite...")
		if err := db.MigrateFromBoltDB(); err != nil {
			log.Fatal("Database migration failed: %v", err)
		}
		migratedFromBoltDB = true
		// Migration succeeded (MigrateFromBoltDB already logged the completion);
		// now initialize SQLite normally.
		if err := db.Open(); err != nil {
			// If SQLite initialization fails after migration, restore the backup
			// so the migration can be retried on next startup.
			confPath := conf.GetEnvironmentConfig().Config
			backupPath := filepath.Join(confPath, "bolt.db.bak")
			if _, e := os.Stat(backupPath); e == nil {
				if renameErr := os.Rename(backupPath, filepath.Join(confPath, "bolt.db")); renameErr == nil {
					log.Warn("Restored bolt.db from bolt.db.bak due to SQLite initialization failure")
				}
			}
			log.Fatal("Failed to initialize SQLite after migration: %v", err)
		}
	} else if err != nil {
		log.Fatal("Failed to initialize database: %v", err)
	}

	//db
	dbPath := filepath.Join(conf.GetEnvironmentConfig().Config, "bolt.db")
	if _, e := os.Lstat(dbPath); os.IsNotExist(e) {
		// If we just migrated from BoltDB to SQLite, the data is already in
		// SQLite. Do NOT run initDBValue() or v4 migration, as that would
		// overwrite or clear the freshly migrated data.
		if !migratedFromBoltDB {
			// db.IsNewDB is true only when v2raya.db was just created by Open()
			// (i.e., it did not exist before). In that case, we need to initialize
			// the default configuration values.
			// If db.IsNewDB is false, v2raya.db already existed (from a previous
			// migration or a previous startup), so we skip initialization.
			if db.IsNewDB {
				// On Windows, v2rayA v4 was never available, so there is no v4
				// data to migrate. Calling any v4 library function here would
				// trigger the v4 library to initialize with its Linux-default
				// config path (/etc/v2raya), which on Windows resolves to
				// \etc\v2raya on the current drive root and causes the library
				// to create that directory and a boltv4.db file there.
				// Skip the v4 migration check entirely on Windows.
				if runtime.GOOS != "windows" && !configurev4.IsConfigureNotExists() {
					// There is different format in server and subscription.
					// So we keep other content and reimport servers and subscriptions.
					log.Warn("Migrating from v4 to main")
					if err := copyfile.CopyFileContent(filepath.Join(
						confv4.GetEnvironmentConfig().Config,
						"boltv4.db",
					), filepath.Join(
						conf.GetEnvironmentConfig().Config,
						"bolt.db",
					)); err != nil {
						log.Fatal("Failed to copy boltv4.db to bolt.db: %v", err)
					}

					// clear connects of outbounds
					for _, out := range configure.GetOutbounds() {
						_ = configure.ClearConnects(out)
					}
					var indexes []int
					for i := 0; i < configurev4.GetLenServers(); i++ {
						indexes = append(indexes, i)
					}
					_ = configure.RemoveServers(indexes)

					indexes = nil
					for i := 0; i < configurev4.GetLenSubscriptions(); i++ {
						indexes = append(indexes, i)
					}
					_ = configure.RemoveSubscriptions(indexes)

					// migrate servers and subscriptions
					t := touchv4.GenerateTouch()
					subs := configurev4.GetSubscriptionsV2()
					for _, sub := range subs {
						log.Info("Importing subscription: %v", sub.Address)
						if e := service.Import(sub.Address, nil); e != nil {
							log.Warn("Failed to migrate subscription: %v", sub.Address)
						}
					}
					for iSvr := range t.Servers {
						if addr, e := servicev4.GetSharingAddress(&configurev4.Which{
							TYPE: configurev4.ServerType,
							ID:   iSvr + 1,
						}); e == nil {
							if e := service.Import(addr, nil); e != nil {
								log.Warn("Failed to migrate server: %v", addr)
							}
						}
					}

					log.Warn("Migration is done")
				} else {
					initDBValue()
				}
			}
		}

		// ensure the default "proxy" outbound group exists
		if err := configure.InitDefaultOutbound(); err != nil {
			log.Warn("initDefaultOutbound: %v", err)
		}
	}

	if len(configure.GetTproxyWhiteIpGroups().CountryCodes) == 0 {
		configure.SetTproxyWhiteIpGroups([]string{"PRIVATE"}, []string{})
	}

	// check if config.json exists
	if _, err := os.Stat(asset.GetV2rayConfigPath()); err != nil {
		// if not exists, create one. This mostly happens when mounting a volume in docker mode and it covers /etc/v2ray.
		t := v2ray.Template{}
		_ = v2ray.WriteV2rayConfig(t.ToConfigBytes())
	}

	// first determine if v2ray exists
	if _, err := where.GetV2rayBinPath(); err == nil {
		// check if geoip, geosite exist
		if !asset.DoesV2rayAssetExist("geoip.dat") || !asset.DoesV2rayAssetExist("geosite.dat") {
			log.Alert("downloading missing geoip.dat and geosite.dat")
			var l net.Listener
			if l, err = net.Listen("tcp", conf.GetEnvironmentConfig().Address); err != nil {
				log.Fatal("net.Listen: %v", err)
			}
			e := gin.New()
			e.GET("/", func(c *gin.Context) {
				c.Header("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate")
				c.Header("Pragma", "no-cache")
				c.Header("Expires", "0")
				c.String(200, "Downloading missing geoip.dat and geosite.dat; refresh the page later.")
			})
			go e.RunListener(l)
			if !asset.DoesV2rayAssetExist("geoip.dat") {
				err := dat.UpdateLocalGeoIP()
				if err != nil {
					log.Fatal("UpdateLocalGeoIP: %v", err)
				}
			}
			if !asset.DoesV2rayAssetExist("geosite.dat") {
				err = dat.UpdateLocalGeoSite()
				if err != nil {
					log.Fatal("UpdateLocalGeoSite: %v", err)
				}
			}
			if l != nil {
				l.Close()
			}
			log.Alert("geoip.dat and geosite.dat are ready")
		}
	}
}
