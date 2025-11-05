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
            <b-field :label="$t('tproxyWhiteIpGroups.formName1')">
                <b-select multiple v-model="countryCodes" expanded>
                    <option value="CN">{{ $t("tproxyWhiteIpGroups.cn") }}</option>
                    <option value="PRIVATE">{{ $t("tproxyWhiteIpGroups.private") }}</option>
                    <option value="US">{{ $t("tproxyWhiteIpGroups.us") }}</option>
                    <option value="CLOUDFLARE">{{ $t("tproxyWhiteIpGroups.cloudflare") }}</option>
                </b-select>
            </b-field>
            <b-field :label="$t('tproxyWhiteIpGroups.formName2')">
                <b-input v-model="customIps" type="textarea" :placeholder="$t('tproxyWhiteIpGroups.formPlaceholder2')"
                    custom-class="full-min-height horizon-scroll code-font" />
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
        countryCodes: [],
        customIps: "",
    }),
    created() {
        this.$axios({
            url: apiRoot + "/tproxyWhiteIpGroups",
        }).then((res) => {
            handleResponse(res, this, () => {
                if (res.data.data.countryCodes) {
                    this.countryCodes = res.data.data.countryCodes;
                }
                if (res.data.data.customIps) {
                    this.customIps = res.data.data.customIps.join("\n");
                }
            });
        });
    },
    methods: {
        validateCIDRArray(arr) {
            const ipv4Cidr = /^(?:25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d)){3}\/(?:[0-9]|[12]\d|3[0-2])$/;

            const ipv6Cidr = /^(?:(?:[A-Fa-f0-9]{1,4}:){7}[A-Fa-f0-9]{1,4}|(?:[A-Fa-f0-9]{1,4}:){1,7}:|(?:[A-Fa-f0-9]{1,4}:){1,6}:[A-Fa-f0-9]{1,4}|(?:[A-Fa-f0-9]{1,4}:){1,5}(?::[A-Fa-f0-9]{1,4}){1,2}|(?:[A-Fa-f0-9]{1,4}:){1,4}(?::[A-Fa-f0-9]{1,4}){1,3}|(?:[A-Fa-f0-9]{1,4}:){1,3}(?::[A-Fa-f0-9]{1,4}){1,4}|(?:[A-Fa-f0-9]{1,4}:){1,2}(?::[A-Fa-f0-9]{1,4}){1,5}|[A-Fa-f0-9]{1,4}:(?:(?::[A-Fa-f0-9]{1,4}){1,6})|:(?:(?::[A-Fa-f0-9]{1,4}){1,7}|:))\/(?:12[0-8]|1[01]\d|[1-9]?\d)$/;

            const invalid = arr
                .map(s => (typeof s === 'string' ? s.trim() : ''))
                .filter(s => s.length === 0 || !(ipv4Cidr.test(s) || ipv6Cidr.test(s)));
            return invalid.length === 0;
        },
        handleClickSubmit() {
            if (!this.validateCIDRArray(this.customIps.split("\n").filter(line => line.trim() !== ''))) {
                this.$buefy.toast.open({
                    message: this.$t("tproxyWhiteIpGroups.invalidCustomIps"),
                    type: "is-danger",
                    position: "is-top",
                    queue: false,
                    duration: 10000,
                });
                return
            }
            this.$axios({
                url: apiRoot + "/tproxyWhiteIpGroups",
                method: "put",
                data: {
                    countryCodes: this.countryCodes.length ? this.countryCodes : ['NONE'],
                    customIps: this.customIps.split("\n").filter(line => line.trim() !== ''),
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
