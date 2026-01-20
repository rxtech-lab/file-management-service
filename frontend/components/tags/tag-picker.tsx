"use client";

import { useState, useEffect } from "react";
import { Check, ChevronsUpDown, Plus } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { listTagsAction, createTagAction } from "@/lib/actions/tag-actions";
import type { Tag } from "@/lib/api/types";

interface TagPickerProps {
  selectedTagIds: number[];
  onSelect: (tag: Tag) => void;
  placeholder?: string;
}

export function TagPicker({
  selectedTagIds,
  onSelect,
  placeholder = "Add tag...",
}: TagPickerProps) {
  const [open, setOpen] = useState(false);
  const [tags, setTags] = useState<Tag[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [isCreating, setIsCreating] = useState(false);

  useEffect(() => {
    async function fetchTags() {
      setIsLoading(true);
      try {
        const result = await listTagsAction();
        if (result.success && result.data) {
          setTags(result.data.data || []);
        }
      } finally {
        setIsLoading(false);
      }
    }
    fetchTags();
  }, []);

  const filteredTags = tags.filter(
    (tag) =>
      !selectedTagIds.includes(tag.id) &&
      tag.name.toLowerCase().includes(search.toLowerCase()),
  );

  const exactMatch = tags.find(
    (tag) => tag.name.toLowerCase() === search.toLowerCase(),
  );

  const handleCreateTag = async () => {
    if (!search.trim() || exactMatch) return;

    setIsCreating(true);
    try {
      const result = await createTagAction({
        name: search.trim(),
        color: "#3b82f6",
      });
      if (result.success && result.data) {
        setTags((prev) => [...prev, result.data!]);
        onSelect(result.data);
        setSearch("");
        setOpen(false);
      }
    } finally {
      setIsCreating(false);
    }
  };

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className="justify-between"
          size="sm"
        >
          {placeholder}
          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[200px] p-0" align="start">
        <Command>
          <CommandInput
            placeholder="Search tags..."
            value={search}
            onValueChange={setSearch}
          />
          <CommandList>
            {isLoading ? (
              <div className="p-2 text-center text-sm text-muted-foreground">
                Loading...
              </div>
            ) : (
              <>
                <CommandEmpty>
                  {search.trim() && !exactMatch ? (
                    <button
                      className="flex items-center gap-2 w-full px-2 py-1.5 text-sm hover:bg-accent rounded"
                      onClick={handleCreateTag}
                      disabled={isCreating}
                    >
                      <Plus className="h-4 w-4" />
                      Create &quot;{search}&quot;
                    </button>
                  ) : (
                    "No tags found"
                  )}
                </CommandEmpty>
                <CommandGroup>
                  {filteredTags.map((tag) => (
                    <CommandItem
                      key={tag.id}
                      value={tag.name}
                      onSelect={() => {
                        onSelect(tag);
                        setSearch("");
                        setOpen(false);
                      }}
                    >
                      <div
                        className="w-3 h-3 rounded-full mr-2"
                        style={{ backgroundColor: tag.color || "#3b82f6" }}
                      />
                      {tag.name}
                    </CommandItem>
                  ))}
                </CommandGroup>
              </>
            )}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
