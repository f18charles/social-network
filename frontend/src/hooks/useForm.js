import { useState } from 'react'

export const useForm = (initialState = {}) => {
    const [formData, setFormData] = useState(initialState)
    const [errors, setErrors] = useState({})

    const handleChange = (e) => {
        const { name, value } = e.target
        setFormData(prev => ({ ...prev, [name]: value }))
        // Clear error for this field
        if (errors[name]) {
            setErrors(prev => ({ ...prev, [name]: '' }))
        }
    }

    const resetForm = () => {
        setFormData(initialState)
        setErrors({})
    }

    const setFieldError = (field, message) => {
        setErrors(prev => ({ ...prev, [field]: message }))
    }

    return {
        formData,
        errors,
        setErrors,
        handleChange,
        resetForm,
        setFieldError
    }
}