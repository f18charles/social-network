import { forwardRef } from 'react'

export const FormInput = forwardRef(({ 
    label, 
    name, 
    type = 'text', 
    value, 
    onChange, 
    onFocus, 
    onBlur,
    error,
    ...props 
}, ref) => {
    return (
        <div className="form-group">
            <label htmlFor={name}>{label}</label>
            <input
                ref={ref}
                type={type}
                id={name}
                name={name}
                value={value}
                onChange={onChange}
                onFocus={onFocus}
                onBlur={onBlur}
                className={error ? 'input-error' : ''}
                {...props}
            />
            {error && <small className="error-text">{error}</small>}
        </div>
    )
})

FormInput.displayName = 'FormInput'