import { DefineFunction, Manifest } from "deno-slack-sdk/mod.ts";

const ReverseFunction = DefineFunction("reverse", {
  "title": "Reverse",
  "description": "Takes a string and reverses it",
});

// Add a user-defined object that has the same name key as `Manifest`
const obj = {
  "name": "random user defined value"
};

export default Manifest({
  "name": "vibrant-butterfly-1234",
  "description": "Reverse a string",
  // "runtime_environment": "slack",
  "runtime": "deno1.x",
  "icon": "assets/icon.png",
  "functions": [ReverseFunction],
  "outgoingDomains": [],
  "botScopes": ["commands", "chat:write", "chat:write.public"],
});
