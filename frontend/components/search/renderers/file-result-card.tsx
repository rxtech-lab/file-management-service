"use client";

import { motion } from "framer-motion";
import { ArrowRight } from "lucide-react";
import {
  FileTypeIcon,
  getFileTypeColor,
} from "@/components/files/file-type-icon";
import type { FileType } from "@/lib/api/types";
import { cn } from "@/lib/utils";

interface DisplayFile {
  id: number;
  title: string;
  file_type: FileType;
  mime_type?: string;
  folder_id: number | null;
  folder?: { name: string };
}

interface FileResultCardProps {
  file: DisplayFile;
  description: string;
  onClick: () => void;
}

export function FileResultCard({
  file,
  description,
  onClick,
}: FileResultCardProps) {
  return (
    <motion.button
      whileHover={{ scale: 1.02, x: 4 }}
      whileTap={{ scale: 0.98 }}
      onClick={onClick}
      className="w-full text-left p-3 rounded-xl border bg-card hover:bg-accent/50 transition-colors group"
    >
      <div className="flex items-start gap-3">
        <div
          className={cn(
            "flex-shrink-0 mt-0.5",
            getFileTypeColor(file.file_type, file.mime_type)
          )}
        >
          <FileTypeIcon
            fileType={file.file_type}
            mimeType={file.mime_type}
            className="h-5 w-5"
          />
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <p className="font-medium text-sm truncate">{file.title}</p>
            <ArrowRight className="h-4 w-4 opacity-0 group-hover:opacity-100 transition-opacity text-muted-foreground" />
          </div>
          {description && (
            <p className="text-xs text-muted-foreground line-clamp-2 mt-1">
              {description}
            </p>
          )}
          {file.folder && (
            <p className="text-xs text-muted-foreground/70 mt-1">
              in {file.folder.name}
            </p>
          )}
        </div>
      </div>
    </motion.button>
  );
}
