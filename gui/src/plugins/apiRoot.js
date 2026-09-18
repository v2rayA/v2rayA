// apiRoot is read as a bare global throughout the interface: "<backend>/api",
// where the backend address lives in localStorage and can change at runtime
// (the address dialog). It used to be a Vite `define` holding a template
// literal, which Vite 5+ rejects — a define must be a literal — so it is a
// getter on window now, evaluated on every read exactly as the define was.
Object.defineProperty(window, "apiRoot", {
  configurable: true,
  get() {
    return `${localStorage["backendAddress"]}/api`;
  },
});
