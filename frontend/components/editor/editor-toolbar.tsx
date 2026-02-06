"use client";

import { Button } from "@/components/ui/button";
import type { SaveStatus } from "@/hooks/use-auto-save";
import type { ToolbarMenuGroup } from "@/lib/plugins/types";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  FolderInput,
  Cloud,
  CloudOff,
  Loader2,
} from "lucide-react";
import type { ReactNode } from "react";
import { EditableTitle } from "./editable-title";
import { ToolbarMenu } from "./toolbar-menu";

interface EditorToolbarProps {
  title: string;
  onTitleChange: (title: string) => Promise<void>;
  menus: ToolbarMenuGroup[];
  saveStatus: SaveStatus;
  lastSavedAt: Date | null;
  saveError: string | null;
  onMoveFile: () => void;
  extraContent?: ReactNode;
  leftContent?: ReactNode;
}

function SaveStatusIcon({
  status,
  lastSavedAt,
  error,
}: {
  status: SaveStatus;
  lastSavedAt: Date | null;
  error: string | null;
}) {
  if (status === "saving") {
    return (
      <Tooltip>
        <TooltipTrigger asChild>
          <div className="p-1.5 rounded-full hover:bg-accent">
            <Loader2 className="h-5 w-5 text-muted-foreground animate-spin" />
          </div>
        </TooltipTrigger>
        <TooltipContent>Saving...</TooltipContent>
      </Tooltip>
    );
  }

  if (status === "error") {
    return (
      <Tooltip>
        <TooltipTrigger asChild>
          <div className="p-1.5 rounded-full hover:bg-accent">
            <CloudOff className="h-5 w-5 text-destructive" />
          </div>
        </TooltipTrigger>
        <TooltipContent>{error || "Save failed"}</TooltipContent>
      </Tooltip>
    );
  }

  if (status === "saved" || lastSavedAt) {
    return (
      <Tooltip>
        <TooltipTrigger asChild>
          <div className="p-1.5 rounded-full hover:bg-accent">
            <Cloud className="h-5 w-5 text-muted-foreground" />
          </div>
        </TooltipTrigger>
        <TooltipContent>
          {lastSavedAt
            ? `Saved at ${lastSavedAt.toLocaleTimeString()}`
            : "Saved"}
        </TooltipContent>
      </Tooltip>
    );
  }

  return null;
}

export function EditorToolbar({
  title,
  onTitleChange,
  menus,
  saveStatus,
  lastSavedAt,
  saveError,
  onMoveFile,
  extraContent,
  leftContent,
}: EditorToolbarProps) {
  return (
    <div className="flex flex-row items-center gap-2 px-3 pt-2 border-b bg-background">
      {leftContent}
      <div className="min-w-0 flex-1">
        {/* Row 1: Title + icon actions */}
        <div className="flex items-center gap-0.5">
          <EditableTitle
            value={title}
            onSave={onTitleChange}
            className="min-w-0"
          />
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8 shrink-0 rounded-full"
                onClick={onMoveFile}
              >
                <FolderInput className="h-5 w-5 text-muted-foreground" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>Move</TooltipContent>
          </Tooltip>
          <SaveStatusIcon
            status={saveStatus}
            lastSavedAt={lastSavedAt}
            error={saveError}
          />
        </div>

        {/* Row 2: Menu bar */}
        <div className="flex items-center gap-0.5 -ml-1.5">
          {menus.map((menu) => (
            <ToolbarMenu key={menu.id} menu={menu} />
          ))}

          <div className="flex-1" />

          {extraContent}
        </div>
      </div>
    </div>
  );
}
