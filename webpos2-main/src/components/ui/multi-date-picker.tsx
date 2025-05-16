'use client'

import * as React from "react"
import { format } from "date-fns"
import { Check, ChevronsUpDown } from "lucide-react"
import { de } from 'date-fns/locale'

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
} from "@/components/ui/command"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"

interface MultiDatePickerProps {
  dates: Date[]
  setDates: (dates: Date[]) => void
  availableDates: Date[]
  label?: string
}

export function MultiDatePicker({ dates, setDates, availableDates, label }: MultiDatePickerProps) {
  const [open, setOpen] = React.useState(false)

  const formatDate = (date: Date) => {
    return format(date, "EEEE, d. MMMM yy", { locale: de })
  }

  const toggleDate = (date: Date) => {
    const dateStr = format(date, 'yyyy-MM-dd')
    const isSelected = dates.some(d => format(d, 'yyyy-MM-dd') === dateStr)
    
    if (isSelected) {
      setDates(dates.filter(d => format(d, 'yyyy-MM-dd') !== dateStr))
    } else {
      setDates([...dates, date])
    }
  }

  return (
    <div className="flex flex-col gap-2">
      {label && <span className="text-sm text-gray-500">{label}</span>}
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            role="combobox"
            aria-expanded={open}
            className="w-[300px] justify-between"
          >
            {dates.length === 0
              ? "Datum wählen"
              : dates.length === 1
              ? formatDate(dates[0])
              : `${dates.length} Tage ausgewählt`}
            <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-[300px] p-0">
          <Command>
            <CommandInput placeholder="Datum suchen..." />
            <CommandEmpty>Keine Daten gefunden.</CommandEmpty>
            <CommandGroup className="max-h-[300px] overflow-auto">
              {availableDates.map((date) => {
                const dateStr = format(date, 'yyyy-MM-dd')
                const isSelected = dates.some(d => format(d, 'yyyy-MM-dd') === dateStr)
                
                return (
                  <CommandItem
                    key={dateStr}
                    value={formatDate(date)}
                    onSelect={() => toggleDate(date)}
                  >
                    <Check
                      className={cn(
                        "mr-2 h-4 w-4",
                        isSelected ? "opacity-100" : "opacity-0"
                      )}
                    />
                    {formatDate(date)}
                  </CommandItem>
                )
              })}
            </CommandGroup>
          </Command>
        </PopoverContent>
      </Popover>
    </div>
  )
}
