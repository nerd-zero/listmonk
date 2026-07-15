import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";
import { toast } from "sonner";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export async function copyToClipboard(value: string, label = "Copied") {
  try {
    await navigator.clipboard.writeText(value);
    toast.success(label);
  } catch {
    toast.error("Couldn't copy — select and copy it manually");
  }
}
