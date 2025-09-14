export function isValidEMail(email) {
    // Check that email is not empty and is a line
    if (typeof email !== 'string' || !email) {
        return false;
    }
    
    // The main regular expression for verification of email
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    
    // Additional checks
    const maxLength = 254; // Maximum email length by standard
    const minLocalPartLength = 1; // Minimum length of the local part
    const maxLocalPartLength = 64; // Maximum length of the local part
    
    // Checking the length
    if (email.length > maxLength) {
        return false;
    }
    
    // Checking the format by regular expression
    if (!emailRegex.test(email)) {
        return false;
    }
    
    // We divide the email into the local part and domain
    const [localPart, domain] = email.split('@');
    
    // Checking the length of the local part
    if (localPart.length < minLocalPartLength || localPart.length > maxLocalPartLength) {
        return false;
    }
    
    // Check that the domain contains at least one point
    if (!domain.includes('.')) {
        return false;
    }
    
    // Check that the domain does not begin and does not end with a point
    if (domain.startsWith('.') || domain.endsWith('.')) {
        return false;
    }
    
    // Check that there are no two points in a row
    if (email.includes('..')) {
        return false;
    }
    
    return true;
}
