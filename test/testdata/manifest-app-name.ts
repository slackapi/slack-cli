import { DefineFunction } from "slack-cloud-sdk/mod.ts";

const reverse = DefineFunction(
  "reverse",
  {
    "title": "reverse-string"
  }
).export();

export default {
  "_metadata": {
    "major_version": 2
  },
  "display_information": {
    "name": "vibrant-butterfly-1234"
  },
  // This is a comment
  "features": {
    "app_home": {
      "home_tab_enabled": false,
      "messages_tab_enabled": false,
      "messages_tab_read_only_enabled": false
    },
    "bot_user": {
      "display_name": "vibrant-butterfly-1234"
    }
  },
  "functions": {
    reverse
  },
  "outgoing_domains": []
}
