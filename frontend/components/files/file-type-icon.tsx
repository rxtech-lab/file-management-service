import {
  FileText,
  Image,
  Music,
  Video,
  FileSpreadsheet,
  File,
  FileArchive,
  FileCode,
  Presentation,
} from "lucide-react";
import type { FileType } from "@/lib/api/types";

interface FileTypeIconProps {
  fileType?: FileType;
  mimeType?: string;
  className?: string;
}

export function FileTypeIcon({
  fileType,
  mimeType,
  className = "h-8 w-8",
}: FileTypeIconProps) {
  // First check fileType enum
  if (fileType) {
    switch (fileType) {
      case "photo":
        return <Image className={className} />;
      case "video":
        return <Video className={className} />;
      case "music":
        return <Music className={className} />;
      case "document":
      case "invoice":
        return <FileText className={className} />;
    }
  }

  // Fall back to MIME type detection
  if (mimeType) {
    if (mimeType.startsWith("image/")) {
      return <Image className={className} />;
    }
    if (mimeType.startsWith("video/")) {
      return <Video className={className} />;
    }
    if (mimeType.startsWith("audio/")) {
      return <Music className={className} />;
    }
    if (mimeType === "application/pdf") {
      return <FileText className={className} />;
    }
    if (
      mimeType.includes("spreadsheet") ||
      mimeType.includes("excel") ||
      mimeType === "text/csv"
    ) {
      return <FileSpreadsheet className={className} />;
    }
    if (mimeType.includes("presentation") || mimeType.includes("powerpoint")) {
      return <Presentation className={className} />;
    }
    if (
      mimeType.includes("zip") ||
      mimeType.includes("tar") ||
      mimeType.includes("compressed") ||
      mimeType.includes("archive")
    ) {
      return <FileArchive className={className} />;
    }
    if (
      mimeType.includes("javascript") ||
      mimeType.includes("json") ||
      mimeType.includes("xml") ||
      mimeType.includes("html") ||
      mimeType.includes("css") ||
      mimeType.startsWith("text/x-")
    ) {
      return <FileCode className={className} />;
    }
    if (
      mimeType.startsWith("text/") ||
      mimeType.includes("document") ||
      mimeType.includes("word")
    ) {
      return <FileText className={className} />;
    }
  }

  // Default icon
  return <File className={className} />;
}

export function getFileTypeColor(
  fileType?: FileType,
  mimeType?: string,
): string {
  if (fileType) {
    switch (fileType) {
      case "photo":
        return "text-green-500";
      case "video":
        return "text-purple-500";
      case "music":
        return "text-pink-500";
      case "document":
        return "text-blue-500";
      case "invoice":
        return "text-orange-500";
    }
  }

  if (mimeType) {
    if (mimeType.startsWith("image/")) return "text-green-500";
    if (mimeType.startsWith("video/")) return "text-purple-500";
    if (mimeType.startsWith("audio/")) return "text-pink-500";
    if (mimeType === "application/pdf") return "text-red-500";
    if (mimeType.includes("spreadsheet") || mimeType.includes("excel"))
      return "text-emerald-500";
  }

  return "text-muted-foreground";
}
