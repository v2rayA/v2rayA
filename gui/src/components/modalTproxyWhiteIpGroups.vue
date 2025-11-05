<template>
    <div class="modal-card" style="max-width: 450px; margin: auto">
        <header class="modal-card-head">
            <p class="modal-card-title">
                {{ $t("tproxyWhiteIpGroups.title") }}
            </p>
        </header>
        <section class="modal-card-body">
            <b-message type="is-info" class="after-line-dot5">
                <p>{{ $t("tproxyWhiteIpGroups.messages.0") }}</p>
            </b-message>
            <b-field :label="$t('tproxyWhiteIpGroups.formName')">
                <b-select multiple v-model="list" expanded>
                    <option value="CN">{{ $t("tproxyWhiteIpGroups.cn") }}</option>
                    <option value="PRIVATE">{{ $t("tproxyWhiteIpGroups.private") }}</option>
                    <option value="US">{{ $t("tproxyWhiteIpGroups.us") }}</option>
                    <option value="CLOUDFLARE">{{ $t("tproxyWhiteIpGroups.cloudflare") }}</option>
                </b-select>
            </b-field>
            <b-message type="is-warning" class="after-line-dot5">
                <p>{{ $t("tproxyWhiteIpGroups.messages.1") }}</p>
            </b-message>
        </section>
        <footer class="modal-card-foot flex-end">
            <button class="button" @click="$emit('close')">
                {{ $t("operations.cancel") }}
            </button>
            <button class="button is-primary" @click="handleClickSubmit">
                {{ $t("operations.save") }}
            </button>
        </footer>
    </div>
</template>
<script>
import { handleResponse } from "@/assets/js/utils";

export default {
    name: "modalTproxyWhiteIpGroups",
    data: () => ({
        list: [],
    }),
    created() {
        this.$axios({
            url: apiRoot + "/tproxyWhiteIpGroups",
        }).then((res) => {
            handleResponse(res, this, () => {
                if (res.data.data.list) {
                    if (res.data.data.list.length) {
                        this.list = res.data.data.list;
                    } else {
                        this.list = ['PRIVATE'];
                    }
                } else {
                    this.list = ['PRIVATE'];
                }
            });
        });
    },
    methods: {
        handleClickSubmit() {
            this.$axios({
                url: apiRoot + "/tproxyWhiteIpGroups",
                method: "put",
                data: {
                    list: this.list.length ? this.list : ['NONE'],
                },
            }).then((res) => {
                handleResponse(res, this, () => {
                    this.$emit("close");
                });
            });
        },
    },
};
</script>
