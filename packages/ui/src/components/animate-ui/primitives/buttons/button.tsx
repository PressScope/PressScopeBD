"use client";

import {
  Slot,
  type WithAsChild,
} from "@PressScopeBd/ui/components/animate-ui/primitives/animate/slot";
import { motion, type HTMLMotionProps } from "motion/react";
import * as React from "react";

type ButtonProps = WithAsChild<
  HTMLMotionProps<"button"> & {
    hoverScale?: number;
    tapScale?: number;
  }
>;

function Button({
  hoverScale = 1.05,
  tapScale = 0.95,
  asChild = false,
  ...props
}: ButtonProps) {
  const Component = asChild ? Slot : motion.button;

  return (
    <Component
      whileTap={{ scale: tapScale }}
      whileHover={{ scale: hoverScale }}
      {...props}
    />
  );
}

export { Button, type ButtonProps };
