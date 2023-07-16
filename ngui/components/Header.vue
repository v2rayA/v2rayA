<script lang="ts" setup>

</script>

<template>
  <header class="p-6 flex justify-between">
    <div class="flex">
      <h1 class="text-large font-600 mr-3">
        <NuxtLink to="/">V2RayA</NuxtLink>
      </h1>

      <OperateBoot />
      <OperateOutbound />
    </div>

    <div class="flex space-x-2">
      <NuxtLink to="/setting">{{ $t('common.setting') }}</NuxtLink>
      <NuxtLink to="/log">{{ $t("common.log") }}</NuxtLink>
      <NuxtLink to="/about">{{ $t("common.about") }}</NuxtLink>
    </div>
  </header>
</template>
