import { createInvoicePlugin } from "./invoice-plugin";
import { createYamlEditorPlugin } from "./yaml-editor-plugin";
import type { FileManagementPlugin } from "../types";

/**
 * Get all built-in plugins with their default configurations.
 * Environment variables are read at plugin creation time.
 */
export function getBuiltinPlugins(): FileManagementPlugin[] {
  return [createInvoicePlugin(), createYamlEditorPlugin()];
}

export { createInvoicePlugin, createYamlEditorPlugin };
