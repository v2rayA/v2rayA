# 플래그와 환경 변수

모든 플래그에는 같은 이름의 환경 변수가 있습니다. `--log-level`에 해당하는 환경 변수는 `V2RAYA_LOG_LEVEL`입니다. Linux 서비스 유닛은 `V2RAYA_LOG_FILE`을 설정하고 `/etc/default/v2raya`를 읽습니다. 나머지는 기본값을 유지합니다. 아래 표는 실행 중인 서비스에서 읽어 오며, 실제 적용된 값이 아닌 선언된 기본값을 표시합니다. 기본값이 비어 있으면 시작 시 계산합니다.

## 훅

`--transparent-hook`은 투명 프록시의 설정 및 해제 전후에 실행할 실행 파일을 지정합니다. 인수로 `--transparent-type=<redirect|tproxy|tun|system_proxy>`, `--stage=<pre-start|post-start|pre-stop|post-stop>`, `--v2raya-confdir=<directory>`를 전달합니다. `--core-hook`은 코어 시작 및 중지 전후에 실행하며, `--stage`와 `--v2raya-confdir`를 전달합니다. **자동 라우트**를 끈 `tun`의 라우트 스크립트는 별도 설정인 **라우트 스크립트 구성**에서 지정합니다.

## 디렉터리

| 항목                                   | 위치                                                                                                                                                                                                                                                         |
| -------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 설정과 `v2raya.db`          | `--config`. 기본값은 Linux와 macOS에서 `/etc/v2raya`, Windows 서비스에서 `%ProgramData%\SYSTEM\v2rayA`, lite 모드에서 `~/.config/v2raya`(또는 `%AppData%\v2rayA`)                                                                                          |
| 생성된 코어 설정                  | 설정 디렉터리의 `config.json`                                                                                                                                                                                                                  |
| 규칙 데이터(`geoip.dat`, `geosite.dat`) | `--v2ray-assetsdir`. 지정하지 않으면 XDG 데이터 디렉터리(`/usr/share/v2raya`, `/usr/local/share/v2raya`, `~/.local/share/v2raya`)를 검색하며, 다운로드한 파일은 사용자 데이터 디렉터리에 저장. Windows에서는 설치 프로그램의 `data` 디렉터리 또는 설정 디렉터리 |
| 코어                               | `--v2ray-bin`. 기본값은 `v2raya`와 같은 디렉터리 또는 `PATH`에 지정된 디렉터리에 있는 `v2raya_core`                                                                                                                                                                                            |
| 로그                                   | `--log-file` 또는 `V2RAYA_LOG_FILE`, 없으면 콘솔. Linux 시스템 유닛은 `/var/log/v2raya/v2raya.log`, 사용자 유닛은 `~/.local/state/v2raya/v2raya.log`, Windows 설치 프로그램은 `%TEMP%\v2raya.log` 사용                                                     |
