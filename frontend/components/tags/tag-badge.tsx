"use client";

import { X } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import type { Tag } from "@/lib/api/types";

interface TagBadgeProps {
  tag: Tag;
  onRemove?: () => void;
  className?: string;
}

export function TagBadge({ tag, onRemove, className }: TagBadgeProps) {
  const backgroundColor = tag.color || "#6b7280";

  return (
    <Badge
      variant="secondary"
      className={className}
      style={{
        backgroundColor: `${backgroundColor}20`,
        borderColor: backgroundColor,
        color: backgroundColor,
      }}
    >
      <span className="truncate max-w-[100px]">{tag.name}</span>
      {onRemove && (
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation();
            onRemove();
          }}
          className="ml-1 rounded-full hover:bg-black/10 p-0.5"
        >
          <X className="h-3 w-3" />
        </button>
      )}
    </Badge>
  );
}
