import type { InputHTMLAttributes } from 'react'

function Input({
  className = '',
  ...props
}: InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      className={`rounded-xl border border-stone-300 bg-white px-3 py-2 text-sm text-stone-800 placeholder:text-stone-400 focus:border-teal-500 focus:ring-1 focus:ring-teal-500 focus:outline-none ${className}`}
      {...props}
    />
  )
}

export default Input
