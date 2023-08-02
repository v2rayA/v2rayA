import Driver from "driver.js";
import "driver.js/dist/driver.min.css";

function nextStep(that, driver, options, step) {
  driver.preventMove();
  driver.reset(true);
  const newDriver = new Driver(options);
  that.$nextTick(() => {
    if (!("offset" in step)) {
      const e = document.querySelector(step.element);
      switch (step.popover.position) {
        case "left":
        case "right":
          step.offset = e.clientHeight / 2;
          break;
        case "top":
        case "bottom":
          step.offset = e.clientWidth / 2;
          break;
      }
      console.log(e, step.offset);
    }
    newDriver.defineSteps([step]);
    newDriver.start();
  });
  return newDriver;
}

export default {
  methods: {
    driver() {
      if (
        this.tableData.subscriptions.length ||
        this.tableData.servers.length ||
        localStorage["drove"] === "true"
      ) {
        return;
      }
      const that = this;
      const options = {
        className: "scoped-class", // className to wrap driver.js popover
        animate: true, // Whether to animate or not
        opacity: 0.75, // Background opacity (0 means only popovers and without overlay)
        padding: 10, // Distance of element from around the edges
        allowClose: false, // Whether the click on overlay should close or not
        overlayClickNext: true, // Whether the click on overlay should move next
        stageBackground: "#ffffff", // Background color for the staged behind highlighted element
        showButtons: false, // Do not show control buttons in footer
        keyboardControl: false, // Allow controlling through keyboard (escape to close, arrow keys to move)
        scrollIntoViewOptions: {}, // We use `scrollIntoView()` when possible, pass here the options for it if you want any
        onHighlightStarted: (Element) => {
          Element.getNode().classList.add("click-through");
        }, // Called when element is about to be highlighted
        onDeselected: (Element) => {
          Element.getNode().classList.remove("click-through");
        },
      };
      let driver = new Driver(options);
      // Define the steps for introduction
      driver.defineSteps([
        {
          element: ".welcome-driver",
          popover: {
            title: that.$t("driver.welcome.0"),
            description: that.$t("driver.welcome.1"),
            position: "left",
          },
          onNext() {
            that.tableData.servers.push({
              id: 1,
              _type: "server",
              name: "ExampleServer ️🇺🇸",
              address: "www.example.com:54321",
              net: "SS(chacha20-ietf-poly1305)",
              pingLatency: "",
            });
            nextStep(that, driver, options, {
              element: ".b-tabs .tabs",
              popover: {
                title: that.$t("driver.tabs.0"),
                description: that.$t("driver.tabs.1"),
                position: "bottom",
              },
            });
          },
        },
      ]);
      driver.start();
    },
  },
};
