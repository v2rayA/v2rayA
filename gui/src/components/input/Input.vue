<template>
  <div class="control" :class="rootClasses">
    <input
      v-if="type !== 'textarea'"
      ref="input"
      class="input"
      :class="[inputClasses, customClass]"
      :type="newType"
      :autocomplete="newAutocomplete"
      :maxlength="maxlength"
      :value="computedValue"
      v-bind="$attrs"
      @input="onInput"
      @blur="onBlur"
      @focus="onFocus"
    />

    <textarea
      v-else
      ref="textarea"
      class="textarea"
      :class="[inputClasses, customClass]"
      :maxlength="maxlength"
      :value="computedValue"
      v-bind="$attrs"
      @input="onInput"
      @blur="onBlur"
      @focus="onFocus"
    ></textarea>

    <b-icon
      v-if="icon"
      class="is-left"
      :icon="icon"
      :pack="iconPack"
      :size="iconSize"
    />

    <b-icon
      v-if="!loading && (passwordReveal || statusTypeIcon)"
      class="is-right"
      :class="{ 'is-clickable': passwordReveal }"
      :icon="passwordReveal ? passwordVisibleIcon : statusTypeIcon"
      :pack="iconPack"
      :size="iconSize"
      :type="!passwordReveal ? statusType : 'is-primary'"
      both
      @click.native="togglePasswordVisibility"
    />

    <small
      v-if="maxlength && hasCounter && type !== 'number'"
      class="help counter"
      :class="{ 'is-invisible': !isFocused }"
    >
      {{ valueLength }} / {{ maxlength }}
    </small>
  </div>
</template>

<script>
import Icon from "buefy/src/components/icon/Icon";
import config from "buefy/src/utils/config";
import FormElementMixin from "buefy/src/utils/FormElementMixin";

export default {
  name: "CusBInput",
  components: {
    [Icon.name]: Icon,
  },
  mixins: [FormElementMixin],
  inheritAttrs: false,
  props: {
    // eslint-disable-next-line vue/require-default-prop
    value: [Number, String],
    type: {
      type: String,
      default: "text",
    },
    passwordReveal: Boolean,
    hasCounter: {
      type: Boolean,
      default: () => config.defaultInputHasCounter,
    },
    customClass: {
      type: String,
      default: "",
    },
  },
  data() {
    return {
      newValue: this.value,
      newType: this.type,
      newAutocomplete: this.autocomplete || config.defaultInputAutocomplete,
      isPasswordVisible: false,
      // eslint-disable-next-line vue/no-reserved-keys
      _elementRef: this.type === "textarea" ? "textarea" : "input",
    };
  },
  computed: {
    computedValue: {
      get() {
        return this.newValue;
      },
      set(value) {
        this.newValue = value;
        this.$emit("input", value);
        !this.isValid && this.checkHtml5Validity();
      },
    },
    rootClasses() {
      return [
        this.iconPosition,
        this.size,
        {
          "is-expanded": this.expanded,
          "is-loading": this.loading,
          "is-clearfix": !this.hasMessage,
        },
      ];
    },
    inputClasses() {
      return [this.statusType, this.size, { "is-rounded": this.rounded }];
    },
    hasIconRight() {
      return this.passwordReveal || this.loading || this.statusTypeIcon;
    },

    /**
     * Position of the icon or if it's both sides.
     */
    // eslint-disable-next-line vue/return-in-computed-property
    iconPosition() {
      if (this.icon && this.hasIconRight) {
        return "has-icons-left has-icons-right";
      } else if (!this.icon && this.hasIconRight) {
        return "has-icons-right";
      } else if (this.icon) {
        return "has-icons-left";
      }
    },

    /**
     * Icon name (MDI) based on the type.
     */
    // eslint-disable-next-line vue/return-in-computed-property
    statusTypeIcon() {
      switch (this.statusType) {
        case "is-success":
          return "check";
        case "is-danger":
          return " iconfont icon-alert";
        case "is-info":
          return "information";
        case "is-warning":
          return "alert";
      }
    },

    /**
     * Check if have any message prop from parent if it's a Field.
     */
    hasMessage() {
      return !!this.statusMessage;
    },

    /**
     * Current password-reveal icon name.
     */
    passwordVisibleIcon() {
      return !this.isPasswordVisible ? "eye" : "eye-off";
    },
    /**
     * Get value length
     */
    valueLength() {
      if (typeof this.computedValue === "string") {
        return this.computedValue.length;
      } else if (typeof this.computedValue === "number") {
        return this.computedValue.toString().length;
      }
      return 0;
    },
  },
  watch: {
    /**
     * When v-model is changed:
     *   1. Set internal value.
     */
    value(value) {
      this.newValue = value;
    },
  },
  methods: {
    /**
     * Toggle the visibility of a password-reveal input
     * by changing the type and focus the input right away.
     */
    togglePasswordVisibility() {
      this.isPasswordVisible = !this.isPasswordVisible;
      this.newType = this.isPasswordVisible ? "text" : "password";

      this.$nextTick(() => {
        this.$refs.input.focus();
      });
    },

    /**
     * Input's 'input' event listener, 'nextTick' is used to prevent event firing
     * before ui update, helps when using masks (Cleavejs and potentially others).
     */
    onInput(event) {
      this.$nextTick(() => {
        if (event.target) {
          this.computedValue = event.target.value;
        }
      });
    },
  },
};
</script>
