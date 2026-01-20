"use client";

import { useCallback } from "react";
import { useFileManagementContext } from "@/lib/plugins/file-management-context";
import type { FileItem } from "@/lib/api/types";
import type { FileOpenHandler } from "@/lib/plugins/types";

export interface UseFileActionReturn {
  /** Check if any plugin can open this file */
  canOpen: (file: FileItem) => boolean;

  /** Get all plugins that can handle this file */
  getOpenHandlers: (file: FileItem) => FileOpenHandler[];

  /** Handle double-click - opens with default/first plugin */
  handleDoubleClick: (file: FileItem) => void;

  /** Handle "Open with" menu item - opens with specific plugin */
  handleOpenWith: (file: FileItem, pluginId: string) => void;

  /** Get the default handler for a file (if any) */
  getDefaultHandler: (file: FileItem) => FileOpenHandler | null;
}

export function useFileAction(): UseFileActionReturn {
  const { canOpen, getOpenHandlers, getDefaultHandler, plugins, callbacks } =
    useFileManagementContext();

  const handleDoubleClick = useCallback(
    (file: FileItem) => {
      const defaultHandler = getDefaultHandler(file);
      if (defaultHandler) {
        defaultHandler.plugin.open(file);
        callbacks?.onOpen?.(file, defaultHandler.plugin.id);
      }
    },
    [getDefaultHandler, callbacks]
  );

  const handleOpenWith = useCallback(
    (file: FileItem, pluginId: string) => {
      const plugin = plugins.find((p) => p.id === pluginId);
      if (plugin && plugin.canOpen(file)) {
        plugin.open(file);
        callbacks?.onOpen?.(file, pluginId);
      } else {
        console.warn(
          `Plugin "${pluginId}" cannot open this file or does not exist`
        );
      }
    },
    [plugins, callbacks]
  );

  return {
    canOpen,
    getOpenHandlers,
    handleDoubleClick,
    handleOpenWith,
    getDefaultHandler,
  };
}
