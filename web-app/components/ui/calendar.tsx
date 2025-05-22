"use client";

import * as React from "react";
import { DayPicker } from "react-day-picker";
import { de } from "date-fns/locale";
import { ChevronLeft, ChevronRight } from "lucide-react";

import { cn } from "@/lib/utils";

export type CalendarProps = React.ComponentProps<typeof DayPicker>;

/**
 * Calendar component stripped of shadcn/ui primary & accent colors.
 * – Always renders **black** (`text-black`) on **white** (`bg-white`).
 * – German localisation (month / weekday names) via `date‑fns/locale/de`.
 * – Weeks start on **Monday** (`weekStartsOn={1}`).
 */
function Calendar({
  className,
  classNames,
  showOutsideDays = true,
  ...props
}: CalendarProps) {
  return (
    <DayPicker
      locale={de}
      weekStartsOn={1}
      showOutsideDays={showOutsideDays}
      className={cn("p-3", className)}
      classNames={{
        // ─── Layout ───────────────────────────────────────────────
        months:
          "flex flex-col sm:flex-row space-y-4 sm:space-x-4 sm:space-y-0 rtl:space-x-reverse",
        month: "space-y-4",
        caption: "flex justify-center pt-1 relative items-center",
        caption_label: "text-sm font-medium text-black",
        table: "w-full border-collapse space-y-1",
        head_row: "flex",
        head_cell:
          "text-black/60 rounded-md w-9 font-normal text-[0.8rem] text-center",
        row: "flex w-full mt-2",

        // ─── Navigation buttons ──────────────────────────────────
        nav: "space-x-1 flex items-center rtl:space-x-reverse",
        nav_button:
          "h-7 w-7 p-0 rounded-md border border-black/30 bg-transparent text-black opacity-75 hover:bg-black hover:text-white hover:opacity-100 focus:outline-none disabled:opacity-40",
        nav_button_previous: "absolute start-1",
        nav_button_next: "absolute end-1",

        // ─── Grid cells ──────────────────────────────────────────
        cell:
          "relative h-9 w-9 p-0 text-center text-sm [&:has([aria-selected].day-range-end)]:rounded-r-md first:[&:has([aria-selected])]:rounded-l-md last:[&:has([aria-selected])]:rounded-r-md focus-within:relative focus-within:z-20",

        // ─── Individual day buttons ─────────────────────────────
        day:
          "h-9 w-9 rounded-md p-0 font-normal text-black hover:bg-black hover:text-white focus:bg-black focus:text-white aria-selected:bg-black aria-selected:text-white",
        day_selected:
          "bg-black text-white hover:bg-black hover:text-white focus:bg-black focus:text-white",
        day_range_end: "day-range-end",
        day_range_middle: "aria-selected:bg-black aria-selected:text-white",
        day_today: "border border-black text-black",
        day_outside:
          "day-outside text-black/40 opacity-60 aria-selected:bg-black aria-selected:text-white",
        day_disabled: "text-black/20 opacity-50",
        day_hidden: "invisible",

        ...classNames,
      }}
      components={{
        IconLeft: ({ className, ...iconProps }) => (
          <ChevronLeft className={cn("h-4 w-4", className)} {...iconProps} />
        ),
        IconRight: ({ className, ...iconProps }) => (
          <ChevronRight className={cn("h-4 w-4", className)} {...iconProps} />
        ),
      }}
      {...props}
    />
  );
}
Calendar.displayName = "Calendar";

export { Calendar };
