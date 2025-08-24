// Mobile menu toggle functionality
function toggleMobileMenu() {
    const mobileMenu = document.getElementById('mobile-menu');
    if (mobileMenu) {
        mobileMenu.classList.toggle('hidden');
    }
}

// Close mobile menu when clicking outside
document.addEventListener('click', function(event) {
    const mobileMenu = document.getElementById('mobile-menu');
    const menuButton = event.target.closest('[onclick="toggleMobileMenu()"]');
    
    if (mobileMenu && !mobileMenu.contains(event.target) && !menuButton) {
        mobileMenu.classList.add('hidden');
    }
});

// Close mobile menu on window resize to desktop size
window.addEventListener('resize', function() {
    const mobileMenu = document.getElementById('mobile-menu');
    if (window.innerWidth >= 768 && mobileMenu) {
        mobileMenu.classList.add('hidden');
    }
});

// Add smooth scrolling for anchor links
document.addEventListener('DOMContentLoaded', function() {
    const links = document.querySelectorAll('a[href^="#"]');
    
    links.forEach(link => {
        link.addEventListener('click', function(e) {
            e.preventDefault();
            
            const targetId = this.getAttribute('href');
            const targetElement = document.querySelector(targetId);
            
            if (targetElement) {
                targetElement.scrollIntoView({
                    behavior: 'smooth'
                });
            }
        });
    });
});

// Form validation helpers
function validateEmail(email) {
    const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return re.test(email);
}

function validatePassword(password) {
    return password.length >= 8;
}

// Show/hide password functionality
function togglePasswordVisibility(inputId, buttonId) {
    const input = document.getElementById(inputId);
    const button = document.getElementById(buttonId);
    
    if (input && button) {
        if (input.type === 'password') {
            input.type = 'text';
            button.innerHTML = '👁️‍🗨️';
        } else {
            input.type = 'password';
            button.innerHTML = '👁️';
        }
    }
}

// Toast notification system
function showToast(message, type = 'info') {
    const toast = document.createElement('div');
    toast.className = `fixed top-4 right-4 p-4 rounded-lg shadow-lg z-50 transition-all duration-300 transform translate-x-full`;
    
    const bgColor = {
        'success': 'bg-green-500',
        'error': 'bg-red-500',
        'warning': 'bg-yellow-500',
        'info': 'bg-blue-500'
    }[type] || 'bg-blue-500';
    
    toast.className += ` ${bgColor} text-white`;
    toast.textContent = message;
    
    document.body.appendChild(toast);
    
    // Animate in
    setTimeout(() => {
        toast.classList.remove('translate-x-full');
    }, 100);
    
    // Auto remove after 5 seconds
    setTimeout(() => {
        toast.classList.add('translate-x-full');
        setTimeout(() => {
            document.body.removeChild(toast);
        }, 300);
    }, 5000);
}