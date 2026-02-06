"use client";

import React, {
  createContext,
  useContext,
  useState,
  useCallback,
  useMemo,
} from "react";
import type { ComponentType } from "react";
import type { FileItem } from "@/lib/api/types";
import type {
  FileManagementPlugin,
  FileManagementContextValue,
  FileOpenHandler,
  FileManagementProviderProps,
  PreviewComponentProps,
} from "./types";

const FileManagementContext = createContext<FileManagementContextValue | null>(
  null
);

export function FileManagementProvider({
  children,
  plugins: initialPlugins = [],
  configs = {},
  callbacks,
}: FileManagementProviderProps) {
  const [plugins, setPlugins] =
    useState<FileManagementPlugin[]>(initialPlugins);
  const [previewFile, setPreviewFile] = useState<FileItem | null>(null);
  const [previewComponent, setPreviewComponent] = useState<ComponentType<PreviewComponentProps> | null>(null);

  const canPreview = useCallback(
    (file: FileItem): boolean => {
      return plugins.some((plugin) => {
        const config = configs[plugin.id];
        if (config?.enabled === false) return false;
        return plugin.onPreview?.(file) !== null && plugin.onPreview?.(file) !== undefined;
      });
    },
    [plugins, configs]
  );

  const openPreview = useCallback(
    (file: FileItem) => {
      for (const plugin of plugins) {
        const config = configs[plugin.id];
        if (config?.enabled === false) continue;
        const component = plugin.onPreview?.(file);
        if (component) {
          setPreviewFile(file);
          setPreviewComponent(() => component);
          return;
        }
      }
    },
    [plugins, configs]
  );

  const openPreviewWith = useCallback(
    (file: FileItem, pluginId: string) => {
      const plugin = plugins.find((p) => p.id === pluginId);
      if (!plugin) return;
      const component = plugin.onPreview?.(file);
      if (component) {
        setPreviewFile(file);
        setPreviewComponent(() => component);
      }
    },
    [plugins]
  );

  const getPreviewHandlers = useCallback(
    (file: FileItem) => {
      return plugins.filter((plugin) => {
        const config = configs[plugin.id];
        if (config?.enabled === false) return false;
        return plugin.onPreview?.(file) != null;
      });
    },
    [plugins, configs]
  );

  const closePreview = useCallback(() => {
    setPreviewFile(null);
    setPreviewComponent(null);
  }, []);

  const registerPlugin = useCallback((plugin: FileManagementPlugin) => {
    setPlugins((prev) => {
      if (prev.some((p) => p.id === plugin.id)) {
        console.warn(`Plugin with id "${plugin.id}" is already registered`);
        return prev;
      }
      return [...prev, plugin];
    });
  }, []);

  const unregisterPlugin = useCallback((pluginId: string) => {
    setPlugins((prev) => prev.filter((p) => p.id !== pluginId));
  }, []);

  const canOpen = useCallback(
    (file: FileItem): boolean => {
      return plugins.some((plugin) => {
        const config = configs[plugin.id];
        if (config?.enabled === false) return false;
        return plugin.canOpen(file);
      });
    },
    [plugins, configs]
  );

  const getOpenHandlers = useCallback(
    (file: FileItem): FileOpenHandler[] => {
      return plugins
        .filter((plugin) => {
          const config = configs[plugin.id];
          if (config?.enabled === false) return false;
          return plugin.canOpen(file);
        })
        .sort((a, b) => (b.priority ?? 0) - (a.priority ?? 0))
        .map((plugin) => ({
          plugin,
          isDefault: plugin.isDefault?.(file) ?? false,
        }));
    },
    [plugins, configs]
  );

  const getDefaultHandler = useCallback(
    (file: FileItem): FileOpenHandler | null => {
      const handlers = getOpenHandlers(file);
      if (handlers.length === 0) return null;

      // First, try to find one marked as default
      const defaultHandler = handlers.find((h) => h.isDefault);
      if (defaultHandler) return defaultHandler;

      // Otherwise, return the first (highest priority)
      return handlers[0];
    },
    [getOpenHandlers]
  );

  const contextValue = useMemo<FileManagementContextValue>(
    () => ({
      plugins,
      registerPlugin,
      unregisterPlugin,
      canOpen,
      getOpenHandlers,
      getDefaultHandler,
      callbacks,
      previewFile,
      previewComponent,
      openPreview,
      openPreviewWith,
      getPreviewHandlers,
      closePreview,
      canPreview,
    }),
    [
      plugins,
      registerPlugin,
      unregisterPlugin,
      canOpen,
      getOpenHandlers,
      getDefaultHandler,
      callbacks,
      previewFile,
      previewComponent,
      openPreview,
      openPreviewWith,
      getPreviewHandlers,
      closePreview,
      canPreview,
    ]
  );

  return (
    <FileManagementContext.Provider value={contextValue}>
      {children}
    </FileManagementContext.Provider>
  );
}

export function useFileManagementContext(): FileManagementContextValue {
  const context = useContext(FileManagementContext);
  if (!context) {
    throw new Error(
      "useFileManagementContext must be used within a FileManagementProvider"
    );
  }
  return context;
}
