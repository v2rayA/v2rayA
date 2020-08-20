// import { parseURL, isIntranet } from "@/assets/js/utils";
console.log("backendPort");
let ba = localStorage.getItem("backendAddress");
if (ba == null) {
  // const u = parseURL(location.href);
  // if (u.host !== "localhost" && u.host !== "local" && isIntranet(u.host)) {
  //   localStorage["backendAddress"] = `${u.protocol}://${u.host}:2017`;
  // } else {
  //   localStorage.setItem("backendAddress", "http://localhost:2017");
  // }
  localStorage.setItem("backendAddress", "");
}
