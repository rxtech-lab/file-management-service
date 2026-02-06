"use client";

import Editor, { type OnMount } from "@monaco-editor/react";
import { useCallback, useRef } from "react";
import type { editor } from "monaco-editor";
import { cn } from "@/lib/utils";

interface MonacoEditorProps {
  value: string;
  onChange?: (value: string) => void;
  language?: string;
  readOnly?: boolean;
  className?: string;
}

export function MonacoEditor({
  value,
  onChange,
  language = "yaml",
  readOnly = false,
  className,
}: MonacoEditorProps) {
  const editorRef = useRef<editor.IStandaloneCodeEditor | null>(null);

  const handleMount: OnMount = useCallback((editor) => {
    editorRef.current = editor;
    editor.focus();
  }, []);

  const handleChange = useCallback(
    (val: string | undefined) => {
      if (val !== undefined && onChange) {
        onChange(val);
      }
    },
    [onChange],
  );

  return (
    <div className={cn("rounded-lg border bg-background overflow-hidden", className)}>
      <Editor
        height="100%"
        language={language}
        value={value}
        onChange={handleChange}
        onMount={handleMount}
        theme="vs-light"
        options={{
          readOnly,
          minimap: { enabled: false },
          wordWrap: "on",
          lineNumbers: "on",
          scrollBeyondLastLine: false,
          fontSize: 14,
          tabSize: 2,
          automaticLayout: true,
          padding: { top: 16, bottom: 16 },
          lineDecorationsWidth: 8,
          renderLineHighlight: "gutter",
          scrollbar: {
            verticalScrollbarSize: 8,
            horizontalScrollbarSize: 8,
          },
        }}
      />
    </div>
  );
}
