import Backbone from "backbone";

// Todo Router
// ----------
module.exports = Backbone.Router.extend({
  routes: {
    "*filter": "setFilter"
  },

  initialize: function({ app }) {
    this.app = app;
  },

  setFilter: function(param) {
    // Set the current filter to be used
    this.app.TodoFilter = param || "";

    // Trigger a collection filter event, causing hiding/unhiding
    // of Todo view items
    this.app.todos.trigger("filter");
  }
});
