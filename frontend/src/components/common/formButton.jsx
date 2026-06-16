export const FormButton = ({ children, onClick, type = 'submit', ...props }) => {
    return (
        <button type={type} onClick={onClick} {...props}>
            {children}
        </button>
    )
}