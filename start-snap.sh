#/bin/bash

sleep 10;
sudo /bin/bash -c "\
snap connect v2raya:firewall-control;\
snap connect v2raya:kernel-module-control;\
snap connect v2raya:network;\
snap connect v2raya:network-bind;\
snap connect v2raya:network-control;\
snap connect v2raya:network-observe;\
snap connect v2raya:v2raya-files;\
snap start v2raya;"
