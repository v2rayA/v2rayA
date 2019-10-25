function locateServer(touch, whichServer) {
  let ind = whichServer.id - 1;
  let sub = whichServer.sub;
  if (whichServer._type === "server") {
    return touch.servers[ind];
  } else if (whichServer._type === "subscription") {
    return touch.subscriptions[sub].servers[ind];
  }
}
export { locateServer };
