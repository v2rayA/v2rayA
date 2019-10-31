import CONST from "./const.js";
function locateServer(touch, whichServer) {
  let ind = whichServer.id - 1;
  let sub = whichServer.sub;
  if (whichServer._type === CONST.ServerType) {
    return touch.servers[ind];
  } else if (whichServer._type === CONST.SubscriptionServerType) {
    return touch.subscriptions[sub].servers[ind];
  }
  return null;
}

function handleResponse(res, that, suc, err) {
  if (res.data.code === "SUCCESS") {
    suc.apply(that);
  } else {
    if (err && err instanceof Function) {
      err.apply(that);
    } else {
      that.$buefy.toast.open({
        message: res.data.message,
        type: "is-warning",
        position: "is-top"
      });
    }
  }
}
export { locateServer, handleResponse };
