"use client";

import { useState, useEffect } from "react";
import { Upload } from "lucide-react";
import { SidebarTrigger } from "@/components/ui/sidebar";
import { Button } from "@/components/ui/button";
import { SearchTrigger } from "@/components/search/search-trigger";
import { SearchCommand } from "@/components/search/search-command";

interface SiteHeaderProps {
  onUploadClick?: () => void;
}

export function SiteHeader({ onUploadClick }: SiteHeaderProps) {
  const [searchOpen, setSearchOpen] = useState(false);

  // Handle Cmd+K keyboard shortcut
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        setSearchOpen(true);
      }
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, []);

  return (
    <>
      <header className="flex h-14 shrink-0 items-center gap-2 border-b px-4">
        <SidebarTrigger className="-ml-1" />

        {/* Search bar - centered */}
        <div className="flex-1 flex justify-center px-4">
          <SearchTrigger onClick={() => setSearchOpen(true)} />
        </div>

        {/* Right side actions */}
        <div className="flex items-center gap-2">
          {onUploadClick && (
            <Button onClick={onUploadClick} size="sm">
              <Upload className="h-4 w-4 mr-2" />
              Upload
            </Button>
          )}
        </div>
      </header>

      <SearchCommand open={searchOpen} onOpenChange={setSearchOpen} />
    </>
  );
}
