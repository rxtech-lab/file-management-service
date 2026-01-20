import type { ReactNode, ComponentType } from "react";
import type { FileItem } from "@/lib/api/types";
import type { LucideIcon } from "lucide-react";

/**
 * Core plugin interface for file management actions.
 * Plugins can handle file opening, and future actions like delete, preview, etc.
 */
export interface FileManagementPlugin {
  /** Unique identifier for the plugin */
  id: string;

  /** Display name shown in menus */
  name: string;

  /** Icon component for menu display */
  icon: LucideIcon | ComponentType<{ className?: string }>;

  /** Priority for ordering (higher = shown first). Default is 0 */
  priority?: number;

  /**
   * Check if this plugin can handle opening the given file.
   * Should be a pure function with no side effects.
   */
  canOpen: (file: FileItem) => boolean;

  /**
   * Open the file using this plugin's handler.
   * Can return a Promise for async operations.
   */
  open: (file: FileItem) => void | Promise<void>;

  /**
   * Whether this is the default handler when double-clicking.
   * Only one plugin should return true for a given file type.
   */
  isDefault?: (file: FileItem) => boolean;
}

/**
 * Result from getOpenHandlers - includes plugin and computed properties
 */
export interface FileOpenHandler {
  plugin: FileManagementPlugin;
  isDefault: boolean;
}

/**
 * Callbacks for file actions - allows parent components to hook into actions
 */
export interface FileActionCallbacks {
  /** Called after a file is opened via any plugin */
  onOpen?: (file: FileItem, pluginId: string) => void;

  /** Called on right-click (future API) */
  onRightClick?: (file: FileItem) => void;

  /** Called on delete action (future API) */
  onDelete?: (file: FileItem) => void;
}

/**
 * Configuration for individual plugins
 */
export interface PluginConfig {
  /** Whether the plugin is enabled */
  enabled?: boolean;

  /** Plugin-specific configuration */
  options?: Record<string, unknown>;
}

/**
 * Context value exposed by FileManagementProvider
 */
export interface FileManagementContextValue {
  /** All registered plugins */
  plugins: FileManagementPlugin[];

  /** Register a new plugin dynamically */
  registerPlugin: (plugin: FileManagementPlugin) => void;

  /** Unregister a plugin by ID */
  unregisterPlugin: (pluginId: string) => void;

  /** Check if any plugin can open the file */
  canOpen: (file: FileItem) => boolean;

  /** Get all plugins that can handle this file, sorted by priority */
  getOpenHandlers: (file: FileItem) => FileOpenHandler[];

  /** Get the default handler for the file (first available or marked as default) */
  getDefaultHandler: (file: FileItem) => FileOpenHandler | null;

  /** Action callbacks */
  callbacks?: FileActionCallbacks;
}

/**
 * Props for the FileManagementProvider component
 */
export interface FileManagementProviderProps {
  children: ReactNode;

  /** Initial plugins to register */
  plugins?: FileManagementPlugin[];

  /** Plugin-specific configurations keyed by plugin ID */
  configs?: Record<string, PluginConfig>;

  /** Callbacks for file actions */
  callbacks?: FileActionCallbacks;
}
