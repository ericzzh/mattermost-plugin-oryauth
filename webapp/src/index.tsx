import { Store, Action } from "redux";

import { GlobalState } from "mattermost-redux/types/store";

import manifest from "./manifest";

// eslint-disable-next-line import/no-unresolved
import { PluginRegistry } from "./types/mattermost-webapp";
import { useEffect } from "react";

import { useLocation } from "react-router-dom";

const PluginRoute = () => {
  const location = useLocation();
  const redirPath = location.pathname.replace("/plug/", "/plugins/");

  useEffect(() => {
    window.location = redirPath + location.search;
  });

  return null;
};
export default class Plugin {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars, @typescript-eslint/no-empty-function
  public async initialize(
    registry: PluginRegistry,
    store: Store<GlobalState, Action<Record<string, unknown>>>
  ) {
    // @see https://developers.mattermost.com/extend/plugins/webapp/reference/
    registry.registerCustomRoute("/", PluginRoute);
  }
}

declare global {
  interface Window {
    registerPlugin(id: string, plugin: Plugin): void;
  }
}

window.registerPlugin(manifest.id, new Plugin());
