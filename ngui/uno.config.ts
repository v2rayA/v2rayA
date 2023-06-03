import {
  defineConfig,
  presetAttributify, presetIcons, presetTagify,
  presetTypography, presetUno,
  transformerDirectives, transformerVariantGroup
} from 'unocss'

export default defineConfig({
  presets: [
    presetUno(),
    presetTypography(),
    presetAttributify({ strict: true }),
    presetIcons({ prefix: '' }),
    presetTagify()
  ],
  transformers: [
    transformerDirectives(),
    transformerVariantGroup()
  ]
})
