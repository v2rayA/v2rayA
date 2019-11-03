<template>
    <div id="app">
        <b-navbar ref="navs" fixed-top shadow type="is-light">
            <template slot="brand">
                <b-navbar-item href="/">
                    <img src="./assets/logo.png" alt="V2RayA" class="logo no-select"/>
                </b-navbar-item>
            </template>
            <template slot="start">
                <b-navbar-item tag="div">
                    V2Ray状态：
                    <b-tag
                            id="statusTag"
                            :type="statusMap[runningState.running]"
                            @mouseenter.native="handleOnStatusMouseEnter"
                            @mouseleave.native="handleOnStatusMouseLeave"
                            @click.native="handleClickStatus"
                    >{{
                        coverStatusText ? coverStatusText : runningState.running
                        }}
                    </b-tag
                    >
                </b-navbar-item>
            </template>

            <template slot="end">
                <!--        <b-navbar-item tag="router-link" to="/node" :active="nav === 'node'">-->
                <!--          <i class="iconfont icon-cloud" style="font-size: 1.4em"></i>-->
                <!--          节点-->
                <!--        </b-navbar-item>-->
                <b-navbar-item tag="a" @click.native="handleClickSetting">
                    <i class="iconfont icon-setting" style="font-size: 1.25em"></i>
                    设置
                </b-navbar-item>
                <b-navbar-item tag="a" @click.native="handleClickAbout">
                    <i class="iconfont icon-heart" style="font-size: 1.25em"></i>
                    关于
                </b-navbar-item>
                <b-dropdown
                        position="is-bottom-left"
                        aria-role="menu"
                        style="margin-right:10px"
                        class="menudropdown"
                >
                    <a slot="trigger" class="navbar-item" role="button">
                        <span class="no-select">mzz2017</span>
                        <i
                                class="iconfont icon-caret-down"
                                style="position: relative; top: 1px; left:2px"
                        ></i>
                    </a>

                    <b-dropdown-item custom aria-role="menuitem">
                        Logged as <b>mzz2017</b>
                    </b-dropdown-item>
                    <hr class="dropdown-divider"/>
                    <b-dropdown-item
                            value="logout"
                            aria-role="menuitem"
                            class="no-select"
                    >
                        <i
                                class="iconfont icon-logout"
                                style="position: relative;top:1px;"
                        ></i>
                        Logout
                    </b-dropdown-item>
                </b-dropdown>
            </template>
        </b-navbar>
        <node v-model="runningState"/>
    </div>
</template>

<script>
    import CONST from "@/assets/js/const";
    import ModalSetting from "@/components/ModalSetting";
    import node from "@/components/node";

    export default {
        components: {node},
        data() {
            return {
                statusMap: {
                    [CONST.INSPECTING_RUNNING]: "is-light",
                    [CONST.NOT_RUNNING]: "is-danger",
                    [CONST.IS_RUNNING]: "is-success"
                },
                coverStatusText: "",
                runningState: {
                    running: CONST.INSPECTING_RUNNING,
                    connectedServer: null,
                    lastConnectedServer: null
                }
            };
        },
        created() {
            this.$axios({
                url: apiRoot + "/version"
            })
                .then(res => {
                    if (res.data.code === "SUCCESS") {
                        this.$buefy.toast.open({
                            message: `V2RayA服务端正在运行于${
                                res.data.data.isInDocker ? "Docker环境中" : "本地"
                                }，Version: ${res.data.data.version}`,
                            type: "is-dark",
                            position: "is-top",
                            duration: 3000
                        });
                    }
                })
                .catch(err => {
                    if (err.message === "Network Error") {
                        console.log("todo", Object.assign({}, err)); //TODO
                        this.$buefy.snackbar.open({
                            message: "未检测到本地V2RayA服务端，请确认服务端是否监听2017端口",
                            type: "is-warning",
                            position: "is-top",
                            actionText: "查看帮助",
                            duration: 60000,
                            onAction: () => {
                                window.open(
                                    "https://github.com/mzz2017/V2RayA#%E4%BD%BF%E7%94%A8",
                                    "_blank"
                                );
                            }
                        });
                    }
                });
        },
        methods: {
            handleOnStatusMouseEnter() {
                if (this.runningState.running === CONST.IS_RUNNING) {
                    this.coverStatusText = "　关闭　";
                } else if (this.runningState.running === CONST.NOT_RUNNING) {
                    this.coverStatusText = "　启动　";
                }
            },
            handleOnStatusMouseLeave() {
                this.coverStatusText = "";
            },
            handleClickSetting() {
                this.$buefy.modal.open({
                    parent: this,
                    component: ModalSetting,
                    hasModalCard: true,
                    canCancel: false
                });
            },
            handleClickAbout() {
                this.$buefy.modal.open({
                    width: 640,
                    content: `
<div class="modal-card" style="margin:auto">
                    <header class="modal-card-head">
                        <p class="modal-card-title">mzz2017 / V2RayA</p>
                    </header>
                    <section class="modal-card-body lazy">
                        <p>V2RayA是V2Ray的一个Web客户端，前端使用Vue.js构建，后端使用Golang构建。</p>
                        <p style="font-size:0.85em;text-indent:1em;color:rgba(0,0,0,0.6)">默认端口：</p>
                        <p style="font-size:0.85em;text-indent:1em;color:rgba(0,0,0,0.6)">2017: V2RayA后端端口</p>
                        <p style="font-size:0.85em;text-indent:1em;color:rgba(0,0,0,0.6)">20170: SOCKS协议</p>
                        <p style="font-size:0.85em;text-indent:1em;color:rgba(0,0,0,0.6)">20171: HTTP协议</p>
                        <p style="font-size:0.85em;text-indent:1em;color:rgba(0,0,0,0.6)">20172: 带PAC的HTTP协议</p>
                        <p>整个项目依赖于Docker，如果你想修改上述端口号，请修改docker参数并重新启动容器。</p>
                        <p>应用不会将任何用户数据保存在云端，所有用户数据存放在docker容器中，当docker容器被清除时配置也将随之消失。</p>
                        <p>在使用中如果发现任何问题，欢迎<a href="https://github.com/mzz2017/V2RayA/issues">提出issue</a>。</p>
                    </section>
                    <footer class="modal-card-foot">
                        <a class="is-link" href="https://github.com/mzz2017/V2RayA" target="_blank">
                          <img class="leave-right" src="https://img.shields.io/github/stars/mzz2017/V2RayA.svg?style=social" alt="stars">
                          <img class="leave-right" src="https://img.shields.io/github/forks/mzz2017/V2RayA.svg?style=social" alt="forks">
                          <img class="leave-right" src="https://img.shields.io/github/watchers/mzz2017/V2RayA.svg?style=social" alt="watchers">
                        </a>
                    </footer>
                </div>
`
                });
            },
            handleClickStatus() {
                if (this.runningState.running === CONST.NOT_RUNNING) {
                    this.$axios({
                        url: apiRoot + "/v2ray",
                        method: "post"
                    }).then(res => {
                        if (res.data.code === "SUCCESS") {
                            Object.assign(this.runningState, {
                                running: CONST.IS_RUNNING,
                                connectedServer: res.data.data.connectedServer,
                                lastConnectedServer: null
                            });
                        } else {
                            this.$buefy.toast.open({
                                message: res.data.message,
                                type: "is-warning",
                                position: "is-top"
                            });
                        }
                    });
                } else if (this.runningState.running === CONST.IS_RUNNING) {
                    this.$axios({
                        url: apiRoot + "/v2ray",
                        method: "delete"
                    }).then(res => {
                        if (res.data.code === "SUCCESS") {
                            Object.assign(this.runningState, {
                                running: CONST.NOT_RUNNING,
                                connectedServer: null,
                                lastConnectedServer: res.data.data.lastConnectedServer
                            });
                            console.log(this.runningState);
                        } else {
                            this.$buefy.toast.open({
                                message: res.data.message,
                                type: "is-warning",
                                position: "is-top"
                            });
                        }
                    });
                }
            }
        }
    };
</script>

<style>
    @import "https://at.alicdn.com/t/font_1467288_0g54ffvhenqu.css";
</style>

<style lang="scss" scoped>
    #app {
        margin: 0;
    }

    .menucontainer {
        padding: 20px;
    }

    .logo {
        min-height: 2.5rem;
    }

    .navbar-item .iconfont {
        margin-right: 0.15em;
    }

    #statusTag:hover {
        cursor: pointer;
    }
</style>

<style lang="scss">
    html {
        &::-webkit-scrollbar {
            // 去掉讨厌的滚动条
            display: none;
        }

        #app {
            height: calc(100vh - 3.25rem);
            overflow-y: auto;
            overflow-scrolling: touch;
            -webkit-overflow-scrolling: touch;
        }
    }

    @media screen and (max-width: 1023px) {
        .dropdown.is-mobile-modal .dropdown-menu {
            // 修复modal模糊问题
            left: 0 !important;
            right: 0 !important;
            margin: auto;
            transform: unset !important;
        }
    }

    .dropdown-item:focus {
        // 不要丑丑的outline
        outline: none !important;
    }

    .no-select {
        user-select: none;
        -webkit-user-drag: none;
    }

    .menudropdown .dropdown-item {
        font-size: 0.8em;
    }

    .leave-right {
        margin-right: 0.6rem;
    }

    .lazy {
        p {
            margin: 0.5em 0;
        }
    }

    a.navbar-item:focus,
    a.navbar-item:focus-within,
    a.navbar-item:hover,
    a.navbar-item.is-active,
    .navbar-link:focus,
    .navbar-link:focus-within,
    .navbar-link:hover,
    .navbar-link.is-active,
    .is-link,
    a {
        $success: #506da4;
        color: $success;
    }

    .icon-loading_ico-copy {
        font-size: 2.5rem;
        color: rgba(0, 0, 0, 0.45);
        animation: loading-rotate 2s infinite linear;
    }
</style>
