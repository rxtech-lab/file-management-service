import { Receipt } from "lucide-react";
import type { FileManagementPlugin } from "../types";
import type { FileItem } from "@/lib/api/types";

/**
 * Creates an invoice plugin that opens invoices in the Invoice Manager app.
 * Requires NEXT_PUBLIC_INVOICE_WEBSITE_URL environment variable to be set.
 */
export function createInvoicePlugin(): FileManagementPlugin {
  const invoiceWebsiteURL = process.env.NEXT_PUBLIC_INVOICE_WEBSITE_URL;

  return {
    id: "invoice-manager",
    name: "Invoice Manager",
    icon: Receipt,
    priority: 100, // High priority for invoice files

    canOpen: (file: FileItem): boolean => {
      if (!invoiceWebsiteURL) return false;
      return file.file_type === "invoice" && !!file.invoice_id;
    },

    open: (file: FileItem): void => {
      if (file.invoice_id && invoiceWebsiteURL) {
        window.open(
          `${invoiceWebsiteURL}/invoices/${file.invoice_id}`,
          "_blank"
        );
      }
    },

    isDefault: (file: FileItem): boolean => {
      return file.file_type === "invoice" && !!file.invoice_id;
    },
  };
}
