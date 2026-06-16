import { useRef } from "react"
import { validateSignIn } from '../../utils/validation'
import { FormButton } from '../common/formButton'
import { FormInput } from '../common/formInput'

export const SignInForm = ({
    formData,
    errors,
    setErrors,
    handleChange,
    resetForm,
    onSwitchToSignUp,
}) => {
    const emailRef = useRef(null)
    const passwordRef = useRef(null)

    const handleFocus = (e) => {
        const label = e.target.previousElementSibling
        if (label) label.className = 'label-selected'
    }

    const handleBlur = (e) => {
        const label = e.target.previousElementSibling
        if (label && e.target.value === '') {
            label.className = ''
        }
    }

    const handleSubmit = (e) => {
        e.preventDefault()

        const validationErrors = validateSignIn(formData)
        if (Object.keys(validationErrors).length > 0) {
            setErrors(validationErrors)
            return
        }

        // TODO: api call
        console.log("signup successful: ", formData)
        alert("sign up successful")

        resetForm()
    }

    return (
        <div className="col-md-6 p-5" id="signUp">
            <h1 className="display-4 text-center">SignIn</h1>
            <form onSubmit={handleSubmit} className="d-flex justify-content-center mt-4">
                <div className="w-75">
                    <FormInput
                        ref={emailRef}
             
    onSuccess           label="Email"
                        name="email"
                        type="email"
                        value={formData.email}
                        onChange={handleChange}
                        onFocus={handleFocus}
                        onBlur={handleBlur}
                        error={errors.email}
                    />
                    <FormInput
                        ref={passwordRef}
                        label="Password"
                        name="password"
                        type="password"
                        value={formData.password}
                        onChange={handleChange}
                        onFocus={handleFocus}
                        onBlur={handleBlur}
                        error={errors.password}
                    />
                    
                    <FormButton>SignIn</FormButton>
                     <div className="d-flex justify-content-between mt-5">
                        <a className="links" href="#" onClick={onSwitchToSignUp}>
                            Create an Account
                        </a>
                    </div>
                </div>
            </form>
        </div>
    )
}