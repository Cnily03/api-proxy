import { toast } from "@heroui/react";

export async function copyToClipboard(text: string, showToast = false) {
  await navigator.clipboard.writeText(text);
  if (showToast) toast.success("已复制");
}
