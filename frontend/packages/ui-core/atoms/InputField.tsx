import { Tooltip } from '@mui/material'
import { ReactElement } from 'react'
import { NumericFormat } from 'react-number-format'

export const StringInputField = (props: {
  initialValue?: string
  className?: string
  placeholder?: string
  inputClassName?: string
  icon?: ReactElement
  onChange: (value: string) => void
}) => {
  return <BaseField {...props} type={'text'} />
}

export const NumberInputField = (props: {
  initialValue?: number
  className?: string
  placeholder?: string
  inputClassName?: string
  icon?: ReactElement
  onChange: (value: string) => void
}) => {
  return <BaseField {...props} type={'number'} />
}

interface props {
  initialValue?: string | number
  className?: string
  placeholder?: string
  inputClassName?: string
  icon?: ReactElement
  type: 'number' | 'text'
  onChange: (value: string) => void
}

const inputText = '!outline-none cursor-pointer focus:cursor-text text-neutral-600 focus:text-neutral-950'

const BaseField = ({ initialValue, inputClassName, placeholder, icon, type, className, onChange }: props) => {
  return (
    <div className={`flex-1 flex items-center p-1 gap-[4px] w-full h-full ${className}`}>
      {icon && (
        <div>
          <Tooltip title={placeholder} placement='top' arrow disableInteractive>
            {icon}
          </Tooltip>
        </div>
      )}

      <input
        className={`${inputText} w-full bg-transparent ${inputClassName}`}
        placeholder={placeholder}
        type={type || 'text'}
        value={initialValue || ''}
        onChange={event => {
          onChange(event.target.value)
        }}
      />
    </div>
  )
}

interface currencyProps {
  className?: string
  value: number
  disabled?: boolean
  onChange?: (value: number) => void
}

export const CurrencyInputField = ({ className, value, onChange, disabled }: currencyProps) => {
  return (
    <NumericFormat
      className={`${className} border border-neutral-300 p-1 w-full text-sm`}
      value={value}
      prefix={'$'}
      disabled={disabled}
      onValueChange={values => {
        if(!onChange) return
        if (values.floatValue) {
          onChange(values.floatValue)
        } else {
          onChange(0)
        }
      }}
    />
  )
}

interface percentageProps {
  className?: string
  value: number
  disabled?: boolean
  onChange: (value: number) => void
}

export const PercentageInputField = ({ className, value, onChange, disabled }: percentageProps) => {
  return (
    <NumericFormat
      className={`${className} border border-neutral-300 p-1 w-full`}
      value={value}
      suffix={'%'}
      disabled={disabled}
      onValueChange={values => {
        if (values.floatValue) {
          onChange(values.floatValue)
        }
      }}
    />
  )
}

interface baseInputProps {
  className?: string
  type: 'number' | 'text'
  placeholder?: string
  value?: string
  onChange: (value: string) => void
}

export const BaseInputField = ({ className, type, placeholder, value, onChange }: baseInputProps) => {
  return (
    <input
      className={`${className} w-full`}
      placeholder={placeholder}
      type={type}
      value={value || ''}
      onChange={event => {
        onChange(event.target.value)
      }}
    />
  )
}
