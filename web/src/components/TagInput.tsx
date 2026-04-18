"use client";

import { Input, Tag, TagGroup } from "@heroui/react";
import { useState } from "react";

interface TagInputProps {
  value: string[];
  onChange: (tags: string[]) => void;
  placeholder?: string;
}

export function TagInput({ value, onChange, placeholder = "输入标签后回车" }: TagInputProps) {
  const [input, setInput] = useState("");

  const addTag = (text: string) => {
    const tag = text.trim();
    if (tag && !value.includes(tag)) {
      onChange([...value, tag]);
    }
    setInput("");
  };

  const removeTag = (key: string) => {
    onChange(value.filter((t) => t !== key));
  };

  return (
    <div className="flex flex-col gap-2">
      {value.length > 0 ? (
        <TagGroup
          selectionMode="none"
          onRemove={(keys) => {
            for (const key of keys) {
              removeTag(String(key));
            }
          }}
        >
          <TagGroup.List>
            {value.map((tag) => (
              <Tag key={tag} id={tag}>
                {tag}
              </Tag>
            ))}
          </TagGroup.List>
        </TagGroup>
      ) : null}
      <Input
        value={input}
        onChange={(e) => setInput(e.target.value)}
        placeholder={placeholder}
        onKeyDown={(e) => {
          if (e.key === "Enter") {
            e.preventDefault();
            addTag(input);
          } else if (e.key === "Backspace" && input === "" && value.length > 0) {
            removeTag(value[value.length - 1]);
          }
        }}
      />
    </div>
  );
}
