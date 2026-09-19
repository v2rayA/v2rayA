import { ref } from "vue";
import {
  deleteTouch,
  getSharingAddress,
  getTouch,
  putSubscription,
} from "@/api";
import { errorText } from "@/api/errors";
import type { TouchSubscription } from "@/api/types";
import { whichOf } from "../nodes/model";

export function useSubscriptions() {
  const subscriptions = ref<TouchSubscription[]>([]);
  const loading = ref(true);
  const loadError = ref("");

  async function sync() {
    loading.value = true;
    loadError.value = "";
    try {
      subscriptions.value = (await getTouch()).touch.subscriptions;
    } catch (err) {
      loadError.value = errorText(err);
      throw err;
    } finally {
      loading.value = false;
    }
  }

  async function update(subscription: TouchSubscription) {
    await putSubscription(whichOf(subscription));
    await sync();
  }

  async function remove(subscription: TouchSubscription) {
    await deleteTouch([whichOf(subscription)]);
    await sync();
  }

  async function sharingLink(subscription: TouchSubscription) {
    return (await getSharingAddress(whichOf(subscription))).sharingAddress;
  }

  return {
    subscriptions,
    loading,
    loadError,
    sync,
    update,
    remove,
    sharingLink,
  };
}
