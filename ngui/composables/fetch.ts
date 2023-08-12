import { nanoid } from 'nanoid'
import { ElMessage } from 'element-plus'
import { createFetch } from '@vueuse/core'

export const useV2Fetch = createFetch({
  baseUrl: `${system.value.api}/api/`,
  combination: 'overwrite',
  options: {
    beforeFetch({ options }) {
      if (user.value.token) {
        options.headers = {
          ...options.headers,
          'Authorization': user.value.token,
          'X-V2raya-Request-Id': nanoid()
        }
      }

      return { options }
    },
    afterFetch({ data }) {
      if (data.code === 'FAIL')
        ElMessage.error({ message: data?.message })

      return { data }
    },
    onFetchError({ error }) {
      if (error)
        ElMessage.error({ message: error?.message })

      return { error }
    }
  }
})
