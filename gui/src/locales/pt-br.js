export default {
  common: {
    outboundSetting: "Configuração de Saída",
    setting: "Configuração",
    about: "Sobre",
    loggedAs: "Logado como <b>{username}</b>",
    v2rayCoreStatus: "Status do v2ray-core",
    checkRunning: "Verificando",
    isRunning: "Executando",
    notRunning: "Pronto",
    notLogin: "Por favor, faça login",
    latest: "Mais recente",
    local: "Local",
    success: "SUCESSO",
    fail: "FALHA",
    message: "Mensagem",
    none: "nenhum",
    optional: "opcional",
    loadBalance: "Balanceamento de Carga",
    log: "Logs",
  },
  welcome: {
    title: "Bem-vindo",
    docker: "O serviço v2rayA está rodando no Docker. Versão: {version}",
    default: "O serviço v2rayA está rodando. Versão: {version}",
    newVersion: "Nova versão detectada: {version}",
    messages: [
      "Não há servidor.",
      "Você pode criar/importar um servidor ou importar uma assinatura.",
    ],
  },
  v2ray: {
    start: "Iniciar",
    stop: "Parar",
  },
  server: {
    name: "Nome do Servidor",
    address: "Endereço do Servidor",
    protocol: "Protocolo",
    latency: "Latência",
    lastSeenTime: "Última vez visto",
    lastTryTime: "Última tentativa",
    messages: {
      notAllowInsecure:
        "De acordo com a documentação do {name}, se você usar {name}, AllowInsecure será proibido.",
      notRecommend:
        "De acordo com a documentação do {name}, se você usar {name}, AllowInsecure não é recomendado.",
    },
  },
  InSecureConfirm: {
    title: "Configuração perigosa detectada",
    message:
      "A configuração definiu o <b>AllowInsecure</b> como verdadeiro. Isso pode causar riscos de segurança. Tem certeza de que deseja continuar?",
    confirm: "Eu sei o que estou fazendo",
    cancel: "cancelar",
  },
  subscription: {
    host: "Host",
    remarks: "Observações",
    timeLastUpdate: "Data e Hora da Última Atualização",
    numberServers: "Número de Servidores",
  },
  operations: {
    name: "Operações",
    update: "Atualizar",
    modify: "Modificar",
    share: "Compartilhar",
    view: "Visualizar",
    delete: "Excluir",
    create: "Criar",
    import: "Importar",
    inBatch: "Em lote",
    connect: "Conectar",
    disconnect: "Desconectar",
    select: "Selecionar",
    login: "Login",
    logout: "Logout",
    configure: "Configurar",
    cancel: "Cancelar",
    saveApply: "Salvar e Aplicar",
    confirm: "Confirmar",
    confirm2: "Cuidadosamente confirmado",
    save: "Salvar",
    copyLink: "COPIAR LINK",
    helpManual: "Ajuda & Manual",
    yes: "Sim",
    no: "Não",
    switchSite: "Mudar para site alternativo",
    addOutbound: "Adicionar uma saída",
  },
  register: {
    title: "Crie uma conta de administrador primeiro",
    messages: [
      "Lembre-se de sua conta de administrador que é importante para o login.",
      "As informações da conta são armazenadas localmente. Nunca enviamos informações para nenhum servidor.",
      "Uma vez que a senha foi esquecida, você pode usar v2raya --reset-password para redefinir.",
    ],
  },
  login: {
    title: "Login",
    username: "Nome de usuário",
    password: "Senha",
  },
  setting: {
    transparentProxy: "Proxy Transparente/Proxy do Sistema",
    transparentType: "Implementação do Proxy Transparente/Proxy do Sistema",
    pacMode: "Modo de Divisão de Tráfego da Porta de Regra",
    preventDnsSpoofing: "Prevenir Falsificação de DNS",
    specialMode: "Modo Especial",
    mux: "Multiplexação",
    autoUpdateSub: "Atualizar Assinaturas Automaticamente",
    autoUpdateGfwlist: "Atualizar GFWList Automaticamente",
    preferModeWhenUpdate: "Modo ao Atualizar Assinaturas e GFWList",
    ipForwardOn: "Encaminhamento de IP",
    portSharingOn: "Compartilhamento de Porta",
    concurrency: "Concorrência",
    options: {
      global: "Não Dividir Tráfego",
      direct: "Direto",
      pac: "Depende da Porta de Regra",
      whitelistCn: "Proxy exceto Sites CN",
      gfwlist: "Proxy apenas GFWList",
      sameAsPacMode: "Modo de Divisão de Tráfego é o Mesmo que a Porta de Regra",
      customRouting: "Roteamento Personalizado",
      antiDnsHijack: "Prevenir apenas sequestro de DNS (rápido)",
      forwardDnsRequest: "Encaminhar Solicitação de DNS",
      doh: "DoH(dns-over-https)",
      default: "Manter Padrão",
      on: "Ligado",
      off: "Desligado",
      updateSubWhenStart: "Atualizar Assinaturas Quando o Serviço Inicia",
      updateSubAtIntervals: "Atualizar Assinaturas Regularmente (Unidade: hora)",
      updateGfwlistWhenStart: "Atualizar GFWList Quando o Serviço Inicia",
      updateGfwlistAtIntervals: "Atualizar GFWList Regularmente (Unidade: hora)",
      dependTransparentMode: "Segue Proxy Transparente/Proxy do Sistema",
      closed: "Desligado",
      advanced: "Configuração Avançada",
      leastPing: "Menor Latência Primeiro",
    },
    messages: {
      gfwlist:
        "Baseado no tempo de modificação do arquivo que às vezes é após a versão mais recente online.",
      transparentProxy:
        "Se o proxy transparente estiver ligado, nenhuma configuração extra é necessária e todo o tráfego TCP passará pelo v2rayA. Fornecendo serviço de proxy para outros computadores e docker como gateway deve fazer a opção 'Compartilhar na LAN' ligada.",
      transparentType:
        "★tproxy: suporta UDP, mas não suporta docker. ★redirect: amigável para docker, mas não suporta UDP e precisa ocupar a porta local 53 para dns anti-poluição.",
      pacMode: `Aqui você pode definir a regra de divisão de tráfego da porta de regra. Por padrão, a "Regra de Divisão de Tráfego" porta é 20172 e protocolo HTTP.`,
      preventDnsSpoofing:
        "★Encaminhar Solicitação de DNS: As solicitações de DNS serão encaminhadas pelo servidor proxy." +
        "★DoH(dns-over-https, v2ray-core: 4.22.0+): DNS sobre HTTPS.",
      specialMode:
        "★supervisor：Monitora a poluição dns, intercepta antecipadamente, usa o mecanismo de sniffing do v2ray-core para prevenir a poluição. ★fakedns：Use a estratégia fakens para acelerar a resolução.",
      tcpFastOpen:
        "Simplifica o processo de handshake do TCP para acelerar o estabelecimento da conexão. Risco de enfatizar as características dos pacotes existe. Pode causar falha na conexão se o seu sistema não suportar.",
      mux:
        "Multiplexando conexões TCP para reduzir o número de handshake, mas isso afetará os casos de uso com alta taxa de transferência, como assistir vídeos, baixar e testar a velocidade. " +
        "Risco de enfatizar as características dos pacotes existe. Suporta apenas vmess agora.",
      confirmEgressPorts: `<p>Você está configurando um proxy transparente em LANs, confirme a lista de permissões da porta de saída.</p>
                          <p>Lista de permissões:</p>
                          <p>TCP: {tcpPorts}</p>
                          <p>UDP: {udpPorts}</p>`,
      grpcShouldWithTls: `gRPC deve estar com TLS`,
      ssPluginImpl:
        "★default: 'transport' para simple-obfs, 'chained' para v2ray-plugin." +
        "★chained: o tráfego shadowsocks será redirecionado para o plugin standalone." +
        "★transport: processado pela camada de transporte do v2ray/xray core diretamente.",
    },
  },
  customAddressPort: {
    title: "Endereço e Portas",
    serviceAddress: "Endereço do Serviço",
    portSocks5: "Porta do SOCKS5",
    portHttp: "Porta do HTTP",
    portSocks5WithPac: "Porta do SOCKS5(com Regra)",
    portHttpWithPac: "Porta do HTTP(com Regra)",
    portVmess: "Porta do VMess(com Regra)",
    portVmessLink: "Link da porta VMess",
    messages: [
      "O endereço do serviço padrão como 0.0.0.0:2017 pode ser alterado definindo a variável de ambiente <code>V2RAYA_ADDRESS</code> e o argumento do comando<code>--address</code>.",
      "Se você iniciar o contêiner docker v2raya com mapeamento de porta em vez de <code>--network host</code>, você pode remapear portas desta maneira.",
      "Não podemos julgar as ocupações de porta no modo docker. Confirme por si mesmo.",
      "Zero significa fechar esta porta.",
    ],
  },
  customRouting: {
    title: "Personalizar Regra de Roteamento",
    defaultRoutingRule: "Regra de Roteamento Padrão",
    sameAsDefaultRule: "o mesmo que a regra padrão",
    appendRule: "Anexar Regra",
    direct: "Direto",
    proxy: "Proxy",
    block: "Bloquear",
    rule: "Regra",
    domainFile: "Arquivo de Domínio",
    typeRule: "Tipo de Regra",
    messages: {
      0: "v2rayA reconhecerá todos os arquivos SiteDat em <b>{V2RayLocationAsset}</b>",
      1: 'Para fazer um arquivo SiteDat por si mesmo: <a href="https://github.com/ToutyRater/V2Ray-SiteDAT">ToutyRater/V2Ray-SiteDAT</a>',
      2: "A seleção múltipla é suportada.",
      noSiteDatFileFound: "Nenhum arquivo siteDat encontrado em {V2RayLocationAsset}",
      emptyRuleNotPermitted: "Regra vazia não é permitida",
    },
  },
  doh: {
    title: "Configurar Servidor DoH",
    dohPriorityList: "Lista de prioridade dos servidores DoH",
    messages: [
      "DoH (DNS sobre HTTPS) pode evitar efetivamente a poluição DNS. Mas alguns provedores nativos de DoH podem ser contaminados às vezes. Além disso, alguns serviços DoH podem ser bloqueados por provedores de rede locais. Por favor, escolha o servidor DoH mais rápido e estável.",
      "Servidores públicos incríveis de DoH na China continental incluem alidns, geekdns, rubyfish, etc",
      "Na área de Taiwan incluem quad101, etc",
      "EUA: cloudflare, dns.google, etc",
      'Lista de verificação：<a href="https://dnscrypt.info/public-servers" target="_blank">public-servers</a>',
      'Além disso, a configuração do serviço DoH em seu próprio servidor nativo é sugerida e bem comportada na maioria dos casos <a href="https://github.com/facebookexperimental/doh-proxy" target="_blank">doh-proxy</a>. Neste caso, é recomendado executar o servidor(doh-proxy/doh-httpproxy) fornecendo serviço e cliente(doh-stub) conectando ao doh.opendns.com ao mesmo tempo e conectá-los em série, porque você dificilmente encontrará um servidor que não está poluído em uma região geralmente contaminada.',
      "Idealmente, coloque uma ou duas linhas acima. A lista será restaurada para o padrão após salvar com conteúdo vazio.",
    ],
  },
dns: {
    title: "Configurar Servidor DNS",
    internalQueryServers: "Servidores de Consulta de Domínio Internos",
    externalQueryServers: "Servidores de Consulta de Domínio Externos",
    messages: [
      '"@:(dns.internalQueryServers)" são projetados para serem usados para pesquisar nomes de domínio na China, enquanto "@:(dns.externalQueryServers)" é usado para pesquisar outros.',
      '"@:(dns.internalQueryServers)" será usado para pesquisar todos os nomes de domínio se "@:(dns.externalQueryServers)" estiver vazio.',
    ],
  },
  egressPortWhitelist: {
    title: "Lista de Portas de Saída",
    tcpPortWhitelist: "Lista de Portas TCP Permitidas",
    udpPortWhitelist: "Lista de Portas UDP Permitidas",
    messages: [
      "Se o v2rayA estiver configurado em um servidor A que está conectado a um servidor proxy B, preste atenção:",
      "O proxy transparente forçará todo o tráfego TCP e UDP a passar pelo servidor proxy B, onde o endereço IP de origem será substituído pelo endereço IP do proxy B. Além disso, se alguns clientes enviarem solicitações ao servidor A que fornece o serviço, eles receberão respostas do endereço IP do seu proxy B de forma estranha, o que é ilegal.",
      "Para resolver isso, precisamos adicionar as portas de serviço à lista de permissões para que não passem pelo proxy. Por exemplo, ssh(22) e v2raya({v2rayaPort}).",
      "Obviamente, se o servidor não fornecer nenhum serviço, você pode pular essa configuração.",
      "Formatação: 22 significa a porta 22, 20170:20172 significa três portas de 20170 a 20172.",
    ],
  },
  configureServer: {
    title: "Configurar Servidor | Servidor",
    servername: "Nome do Servidor",
    port: "Porta",
    forceTLS: "Forçar TLS",
    noObfuscation: "Sem ofuscação",
    httpObfuscation: "Ofuscado como HTTP",
    srtpObfuscation: "Ofuscado como Chamadas de Vídeo (SRTP)",
    utpObfuscation: "Ofuscado como Bittorrent (uTP)",
    wechatVideoObfuscation: "Ofuscado como Chamadas de Vídeo do WeChat",
    dtlsObfuscation: "Ofuscado como Pacotes DTLS1.2",
    wireguardObfuscation: "Ofuscado como Pacotes WireGuard",
    hostObfuscation: "Host",
    pathObfuscation: "Caminho",
    seedObfuscation: "Seed",
    username: "Nome de Usuário",
    password: "Senha",
    origin: "origem",
  },
  configureSubscription: {
    title: "Configurar Assinatura",
  },
  import: {
    message: "Digite um link de servidor ou endereço de assinatura:",
    batchMessage: "Um link de servidor por linha:",
    qrcodeError: "Não foi possível encontrar um código QR válido, tente novamente",
  },
  delete: {
    title: "Confirmar EXCLUSÃO",
    message:
      "Tem certeza de que deseja <b>EXCLUIR</b> esses servidores/assinaturas? Essa ação não pode ser desfeita.",
  },
  latency: {
    message:
      "Os testes de latência costumavam levar um ou vários minutos. Aguarde pacientemente, por favor.",
  },
  version: {
    higherVersionNeeded:
      "Essa operação requer uma versão mais recente do v2rayA do que {version}",
    v2rayInvalid:
      "O arquivo geosite.dat, geoip.dat ou o v2ray-core podem não estar instalados corretamente",
    v2rayNotV5:
      "A versão do v2ray-core não é a versão 5. Use a versão 5 ou faça o downgrade do v2rayA para a versão 1.5",
  },
  about: `<p>O v2rayA é um cliente GUI da V2Ray.</p>
          <p class="about-small">Portas padrão:</p>
          <p class="about-small">2017: porta de serviço do v2rayA</p>
          <p class="about-small">20170: protocolo SOCKS</p>
          <p class="about-small">20171: protocolo HTTP</p>
          <p class="about-small">20172: protocolo HTTP com "Regra de Divisão de Tráfego"</p>
          <p class="about-small">Outras portas:</p>
          <p class="about-small">32345: tproxy, necessário para proxy transparente</p>
          <p>Todos os dados são armazenados localmente, não na nuvem.</p>
          <p>Problemas encontrados durante o uso podem ser relatados em <a href="https://github.com/v2rayA/v2rayA/issues">issues</a>.</p>
          <p>Documentação: <a href="https://v2raya.org">https://v2raya.org</a></p>`,
  axios: {
    messages: {
      optimizeBackend: "Ajustar o endereço do serviço v2rayA?",
      noBackendFound:
        "Não é possível encontrar o v2rayA em {url}. Verifique se o v2rayA está sendo executado nesse endereço.",
      cannotCommunicate: [
        "Não é possível comunicar. Se o seu serviço estiver em execução e as portas estiverem abertas corretamente, a razão pode ser que o navegador atual não permite que sites https acessem recursos http. Você pode tentar usar o Chrome ou alternar para um site http alternativo.",
        "Não é possível comunicar. O Firefox não permite que sites https acessem recursos http. Você pode tentar alternar para sites http alternativos.",
      ],
    },
    urls: {
      usage: "https://github.com/v2rayA/v2rayA/wiki/Usage",
    },
  },
  routingA: {
    messages: ["clique no botão 'Help&Manual' para obter ajuda"],
  },
  outbound: {
    addMessage: "Digite o nome do destino de saída que deseja adicionar:",
    deleteMessage:
      'Tem certeza de que deseja <b>EXCLUIR</b> o destino de saída "{outboundName}"? Essa ação não pode ser desfeita.',
  },
  log: {
    logModalTitle: "Visualizar logs",
    refreshInterval: "Intervalo de atualização",
    seconds: "segundos",
    autoScoll: "Rolagem Automática
  },
};
