import auth_style from '../../styles/auth.css'
import {useForm} from '../../hooks/useForm'
import {useAuth} from '../../hooks/useAuth'
import {SignUpForm, SignInForm, AuthOverlay} from './'

const Auth = () => {
    const signupForm = useForm({email: "", passoword: "", confirm: ""})
    const signinForm = useForm({email:"", password:""})
    const { isMobile, showSignUp, switchToSignUp, switchToSignIn } = useAuth()

    const handleSwitchToSignIn = (e) => {
        e?.preventDefault()
        switchToSignIn()
        if (!isMobile) {
            document.getElementById("overlay").style.transform='translate(0px, -25px)'
        }
    }

    const handleSwitchToSignUp = (e) => {
        e?.preventDefault()
        switchToSignUp()
        if (!isMobile) {
            document.getElementById("overlay").style.transform='translate(550px, -25px)'
        }
    }

    const handleSignupSuccess = () => {
        if (isMobile) {
            switchToSignIn
        } else {
            document.getElementById("overlay").style.transform='translate(0px, -25px'
        }
    }

    return (
        <>
            <style>{auth_style}</style>
            <div className="container p-4">
                <div className="row main mt-5 position-relative">
                    <SignUpForm
                        {...signupForm}
                        onSwitchToSignIn={handleSwitchToSignIn}
                        onSuccess={handleSignupSuccess}
                        style={{
                            display: isMobile && !showSignUp ? 'none' : 'block'
                        }}
                    />
                    <SignInForm
                        {...signinForm}
                        onSwitchToSignUp={handleSwitchToSignUp}
                        style={{
                            display: isMobile && showSignUp ? 'none' : 'block'
                        }}
                    />
                    {!isMobile && <AuthOverlay/>}
                </div>
            </div>
        </>
    )
}

export default Auth