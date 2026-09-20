# 라우팅 규칙

RoutingA는 v2rayA의 규칙 언어입니다. **설정 → RoutingA**에서 편집기를 엽니다. 규칙 포트(`20172`)의 모드가 RoutingA일 때 해당 포트에 규칙이 적용되고, 투명 프록시 정책이 _규칙 포트와 동일_일 때 투명 프록시에도 적용됩니다. RoutingA에 연결된 사용자 지정 인바운드는 별도의 규칙 텍스트를 사용합니다. 아래를 참고하세요.

## 문법

한 줄에 규칙 하나를 작성합니다. 조건을 `&&`로 연결한 뒤 `->`와 아웃바운드를 적습니다. 규칙은 위에서 아래로 검사하며 처음 일치하는 규칙이 적용됩니다. `default:`는 어떤 규칙에도 일치하지 않는 트래픽의 아웃바운드를 지정합니다. `#`으로 시작하는 줄은 주석입니다. v2rayA는 자체 규칙 몇 개를 앞에 추가합니다. 프록시 서버 주소와 Apple 푸시 서비스에는 직접 연결하며, `inboundTag(...)` 조건은 규칙의 일반적인 인바운드 적용 범위를 대체합니다.

```
default: proxy
domain(geosite:category-ads-all) -> block
domain(geosite:cn) -> direct
ip(geoip:private, geoip:cn) -> direct
```

| 조건         | 일치 대상                                                                                                                                                                  |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `domain(...)`     | `domain:`은 도메인과 하위 도메인, `full:`은 정확히 일치하는 도메인, `regexp:`, `geosite:<category>`, `ext:"file.dat:tag"`. `example.com`처럼 접두사 없는 값은 부분 문자열로 일치 여부를 판단 |
| `ip(...)`         | 주소, CIDR(IPv6는 따옴표로 묶음: `ip("2001:db8::/32")`), `geoip:<code>`, `ext:"file.dat:tag"`                                                                             |
| `port(...)`       | 목적지 포트와 범위, `port(80, 443, 1000-2000)`                                                                                                                 |
| `sourcePort(...)` | 출발지 포트                                                                                                                                                             |
| `network(...)`    | `tcp`, `udp`                                                                                                                                                             |
| `protocol(...)`   | 스니핑한 프로토콜: `http`, `tls`, `quic`, `bittorrent`(**인바운드 스니핑**을 켜야 함)                                                                                            |
| `source(...)`     | 출발지 주소와 CIDR                                                                                                                                               |
| `inboundTag(...)` | 연결이 들어온 인바운드                                                                                                                                    |

## 아웃바운드

`proxy`, `direct`, `block`은 기본 제공됩니다. 그 외 프록시 그룹은 연결된 구성원이 있으면 그룹 이름으로 아웃바운드가 됩니다. 연결된 구성원이 없으면 규칙 적용에 실패합니다. v2rayA 외부의 SOCKS 또는 HTTP 서버는 한 번 선언한 뒤 그룹처럼 사용합니다(`ext:` 파일은 규칙 데이터 디렉터리에 넣음).

```
outbound: office = socks(address: 192.168.1.10, port: 1080)
outbound: gateway = http(address: 10.0.0.1, port: 8080, user: 'name', pass: 'secret')
domain(domain: corp.example) -> office
```

## 카테고리

`geosite:` 카테고리는 v2fly의 [domain-list-community](https://github.com/v2fly/domain-list-community)에서, `geoip:` 코드는 [v2fly/geoip](https://github.com/v2fly/geoip)에서 가져옵니다. v2rayA는 처음 시작할 때 없는 파일을 다운로드하며, 패키지에서 자체 빌드를 제공할 수도 있습니다. 자주 사용하는 항목은 `geosite:cn`, `geosite:geolocation-!cn`, `geosite:private`, `geosite:category-ads-all`, `geosite:greatfire`, `geosite:netflix`, `geosite:youtube`, `geosite:telegram`, `geosite:openai`, `geoip:cn`, `geoip:private`, 국가 코드입니다. 속성으로 카테고리 범위를 좁힐 수 있습니다: `geosite:apple@cn`, `geosite:category-games@cn`.

Loyalsoldier 빌드에만 있는 이름(`gfw`, `apple-cn`, `google-cn`, `geoip:telegram`)은 이 파일들에 없습니다. 코어는 `illegal domain rule` 또는 `illegal ip rule` 오류로 해당 이름을 거부합니다. **GFWList 업데이트**는 해당 빌드를 `v2rayA/dist-v2ray-rules-dat`에서 `LoyalsoldierSite.dat`로 다운로드합니다. 이후 규칙에서 `domain(ext:"LoyalsoldierSite.dat:gfw")`로 사용할 수 있습니다.

## 편집기

목록 모드에서는 양식으로 각 규칙을 편집합니다. 텍스트 모드에서는 줄 번호, 구문 강조, 줄별 검사 결과를 표시합니다. 편집기 옆의 문법 안내에서 커서 위치에 예제를 삽입할 수 있으며, **템플릿** 섹션에는 일반적인 정책에 맞는 전체 규칙 세트가 있습니다. 자신의 규칙 위에 삽입하거나 전체 규칙 세트를 대체하세요. 규칙은 텍스트 파일로 가져오고 내보냅니다.

저장하면 실행 중인 코어가 새 규칙을 다시 불러옵니다. 코어가 규칙을 거부하면 v2rayA는 이전 규칙으로 되돌리고 문제가 된 텍스트가 명시된 코어 오류를 표시합니다. 중지된 코어는 다음 시작 시 규칙을 적용합니다.

## 사용자 지정 인바운드

RoutingA에 연결된 사용자 지정 인바운드는 해당 인바운드에만 적용되는 별도의 규칙 텍스트를 사용합니다. 정의가 아닌 규칙만 받습니다. `default:`와 `outbound:` 줄은 무시하며, 어떤 규칙에도 일치하지 않는 트래픽은 양식에서 선택한 그룹으로 보냅니다. 나머지 트래픽의 경로도 직접 지정하려면 `network(tcp, udp) -> direct`처럼 모든 트래픽에 일치하는 규칙으로 텍스트를 끝내세요.
