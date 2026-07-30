import { useState, useEffect } from 'react'

export const useAuth = () => {
    const [isMobile, setIsMobile] = useState(window.innerWidth < 600)
    const [showSignUp, setShowSignUp] = useState(true)

    useEffect(() => {
        const handleResize = () => {
            setIsMobile(window.innerWidth < 600)
        }
        window.addEventListener('resize', handleResize)
        return () => window.removeEventListener('resize', handleResize)
    }, [])

    const switchToSignUp = () => {
        if (isMobile) {
            setShowSignUp(true)
        }
    }

    const switchToSignIn = () => {
        if (isMobile) {
            setShowSignUp(false)
        }
    }

    return {
        isMobile,
        showSignUp,
        setShowSignUp,
        switchToSignUp,
        switchToSignIn
    }
}