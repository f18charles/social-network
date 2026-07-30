export const validateEmail = (email) => {
    if (!email) return "Email is required"
    if (!/\S+@\S+\.\S+/.test(email)) return "Email is invalid"
    return null
}

export const validatePassword = (password) => {
    if (!password) return "Password is required"
    if (password.length < 6) return "Password must be at least 6 characters"
    return null
}

export const validateConfirmPassword = (password, confirm) => {
    if (!confirm) return "Please confirm your password"
    if (password !== confirm) return "Passwords don't match"
    return null
}

export const validateSignUp = (data) => {
    const errors = {}
    
    const emailError = validateEmail(data.email)
    if (emailError) errors.email = emailError
    
    const passwordError = validatePassword(data.password)
    if (passwordError) errors.password = passwordError
    
    const confirmError = validateConfirmPassword(data.password, data.confirm)
    if (confirmError) errors.confirm = confirmError
    
    return errors
}

export const validateSignIn = (data) => {
    const errors = {}
    
    const emailError = validateEmail(data.email)
    if (emailError) errors.email = emailError
    
    const passwordError = validatePassword(data.password)
    if (passwordError) errors.password = passwordError
    
    return errors
}