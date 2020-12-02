import Backbone from "backbone";
// Todo Model
// ----------

// Our basic **Todo** model has `title`, `order`, and `completed` attributes.
module.exports = Backbone.Model.extend({
  // Default attributes for the todo
  // and ensure that each todo created has `title` and `completed` keys.
  defaults: {
    title: "",
    completed: false
  },

  idAttribute: "url",

  url: function() {
    if (this.isNew()) {
      return this.collection.url;
    } else {
      return this.collection.url + this.get("url");
    }
  },

  // Toggle the `completed` state of this todo item.
  toggle: function() {
    this.save(
      {
        completed: !this.get("completed")
      },
      { patch: true }
    );
  }
});
