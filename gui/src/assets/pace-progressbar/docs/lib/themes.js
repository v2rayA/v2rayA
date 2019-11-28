(function() {
  var Color, compileTheme, loadTheme, vm;

  if (typeof module !== "undefined" && module !== null) {
    vm = require('vm');
    Color = require('src/assets/pace-progressbar/docs/lib/color');
  }

  loadTheme = function(name, cb) {
    return $.ajax({
      url: "/pace/templates/pace-theme-" + name + ".tmpl.css",
      success: cb
    });
  };

  compileTheme = function(body, args) {
    if (args == null) {
      args = {};
    }
    return body.replace(/`([\s\S]*?)`/gm, function(match, code) {
      var val;
      if (typeof module !== "undefined" && module !== null) {
        val = vm.runInNewContext(code, {
          args: args,
          Color: Color
        });
      } else {
        Color = window.Color;
        val = eval(code);
      }
      return val;
    });
  };

  if (typeof module !== "undefined" && module !== null) {
    module.exports = {
      compileTheme: compileTheme
    };
  } else {
    window.loadTheme = loadTheme;
    window.compileTheme = compileTheme;
  }

}).call(this);
