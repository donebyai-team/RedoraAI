import React, { FC, useState } from 'react'
import Box from '@mui/material/Box'
import Slider from '@mui/material/Slider'
import { useDebounce } from '../hooks/useDebounce'

type Props = {
  min: number
  max: number
  step?: number
  defaultValue?: number
  onChangeCommitted?: (value: number) => void
}

export const InputSlider: FC<Props> = ({ min, max, step, defaultValue, onChangeCommitted }) => {
  const [value, setValue] = useState(defaultValue || min)
  const debouncedOnChangeCommitted = useDebounce(onChangeCommitted ?? (() => {}), 300)

  const handleSliderChange = (_event: Event, newValue: number | number[]) => {
    setValue(Array.isArray(newValue) ? newValue[0] : newValue)
  }

  const handleSliderCommitted = (_event: unknown, newValue: number | number[]) => {
    setValue(Array.isArray(newValue) ? newValue[0] : newValue)

    // No debounce here since the commit event here is triggered on mouse up
    onChangeCommitted && onChangeCommitted(value)
  }

  const handleInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = boundValueToMinMax(Number(event.target.value), min, max)

    setValue(newValue)
    debouncedOnChangeCommitted(newValue)
  }

  const handleBlur = () => {
    setValue(prev => boundValueToMinMax(prev, min, max))
  }

  return (
    <Box className='flex gap-1 w-full'>
      <Slider
        value={value || min}
        aria-labelledby='input-slider'
        min={min}
        max={max}
        step={step || 1}
        className='flex-1'
        onChange={handleSliderChange}
        onChangeCommitted={handleSliderCommitted}
      />
      <input
        type='number'
        value={value}
        onChange={handleInputChange}
        onBlur={handleBlur}
        className='w-[48px] text-center rounded-md border border-neutral-200 text-sm'
      />
    </Box>
  )
}

function boundValueToMinMax(value: number, min: number, max: number): number {
  if (value < min) {
    return min
  } else if (value > max) {
    return max
  }

  return value
}
