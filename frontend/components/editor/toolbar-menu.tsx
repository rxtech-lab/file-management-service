"use client";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import type { ToolbarMenuGroup } from "@/lib/plugins/types";

interface ToolbarMenuProps {
  menu: ToolbarMenuGroup;
}

function formatShortcut(shortcut: string): string {
  const isMac =
    typeof navigator !== "undefined" &&
    /Mac|iPod|iPhone|iPad/.test(navigator.userAgent);

  return shortcut
    .replace(/Ctrl/gi, isMac ? "\u2318" : "Ctrl")
    .replace(/Shift/gi, isMac ? "\u21E7" : "Shift")
    .replace(/Alt/gi, isMac ? "\u2325" : "Alt")
    .replace(/\+/g, isMac ? "" : "+");
}

export function ToolbarMenu({ menu }: ToolbarMenuProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          size="sm"
          className="h-7 px-2 text-sm font-normal rounded-sm"
        >
          {menu.label}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="min-w-[200px]">
        {menu.items.map((item) => (
          <div key={item.id}>
            {item.separator && <DropdownMenuSeparator />}
            <DropdownMenuItem
              disabled={item.disabled}
              onSelect={() => item.onAction()}
              className="flex items-center justify-between gap-4"
            >
              <div className="flex flex-col">
                <span>{item.title}</span>
                {item.description && (
                  <span className="text-xs text-muted-foreground">
                    {item.description}
                  </span>
                )}
              </div>
              {item.shortcut && (
                <span className="text-xs text-muted-foreground ml-auto shrink-0">
                  {formatShortcut(item.shortcut)}
                </span>
              )}
            </DropdownMenuItem>
          </div>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
