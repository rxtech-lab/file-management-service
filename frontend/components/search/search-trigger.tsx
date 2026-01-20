"use client";

import { Search } from "lucide-react";
import { Button } from "@/components/ui/button";

interface SearchTriggerProps {
  onClick: () => void;
}

export function SearchTrigger({ onClick }: SearchTriggerProps) {
  return (
    <Button
      variant="outline"
      onClick={onClick}
      className="w-full max-w-md justify-start text-muted-foreground"
    >
      <Search className="h-4 w-4 mr-2" />
      <span className="flex-1 text-left">Search files...</span>
      <kbd className="pointer-events-none hidden h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium opacity-100 sm:flex">
        <span className="text-xs">⌘</span>K
      </kbd>
    </Button>
  );
}
