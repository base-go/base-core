import { clsx } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs) {
  return twMerge(clsx(inputs));
}

// Utility for generating component variants
export function createVariants(config) {
  return function(props) {
    const { variants, defaultVariants } = config;
    const mergedProps = { ...defaultVariants, ...props };
    
    let classes = [];
    
    for (const [key, value] of Object.entries(mergedProps)) {
      if (variants[key] && variants[key][value]) {
        classes.push(variants[key][value]);
      }
    }
    
    return cn(...classes);
  };
}