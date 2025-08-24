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

// Modal functionality
function openModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.classList.remove('hidden');
        document.body.classList.add('overflow-hidden');
    }
}

function closeModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.classList.add('hidden');
        document.body.classList.remove('overflow-hidden');
    }
}

// Confirmation dialog functionality
let confirmationCallback = null;

function confirmDelete(userId, userName) {
    const modal = document.getElementById('confirmation-modal');
    const message = document.getElementById('confirmation-message');
    
    if (modal && message) {
        message.textContent = `Are you sure you want to delete ${userName}? This action cannot be undone.`;
        confirmationCallback = () => deleteUser(userId);
        openModal('confirmation-modal');
    }
}

function confirmAction(modalId) {
    if (confirmationCallback) {
        confirmationCallback();
        confirmationCallback = null;
    }
    closeModal(modalId);
}

function confirmDeleteAccount() {
    const modal = document.getElementById('confirmation-modal');
    const message = document.getElementById('confirmation-message');
    
    if (modal && message) {
        message.textContent = 'Are you sure you want to delete your account? This will permanently delete all your data and cannot be undone.';
        confirmationCallback = () => deleteAccount();
        openModal('confirmation-modal');
    }
}

// User management functions
function deleteUser(userId) {
    fetch(`/admin/users/${userId}`, {
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json',
        }
    })
    .then(response => {
        if (response.ok) {
            showToast('User deleted successfully', 'success');
            // Remove the user row from the table
            const userRow = document.querySelector(`tr[data-user-id="${userId}"]`);
            if (userRow) {
                userRow.remove();
            }
        } else {
            showToast('Failed to delete user', 'error');
        }
    })
    .catch(error => {
        console.error('Error:', error);
        showToast('An error occurred while deleting the user', 'error');
    });
}

function deleteAccount() {
    fetch('/profile/delete', {
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json',
        }
    })
    .then(response => {
        if (response.ok) {
            showToast('Account deleted successfully', 'success');
            setTimeout(() => {
                window.location.href = '/';
            }, 2000);
        } else {
            showToast('Failed to delete account', 'error');
        }
    })
    .catch(error => {
        console.error('Error:', error);
        showToast('An error occurred while deleting your account', 'error');
    });
}

// Search functionality
let searchTimeout;

function searchUsers(query) {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => {
        const currentUrl = new URL(window.location);
        if (query.trim()) {
            currentUrl.searchParams.set('search', query);
        } else {
            currentUrl.searchParams.delete('search');
        }
        currentUrl.searchParams.delete('page'); // Reset to first page
        window.location.href = currentUrl.toString();
    }, 500);
}

// Form handling
function handleFormSubmit(formId, successMessage) {
    const form = document.getElementById(formId);
    if (!form) return;
    
    form.addEventListener('submit', function(e) {
        e.preventDefault();
        
        const formData = new FormData(form);
        const submitButton = form.querySelector('button[type="submit"]');
        const originalText = submitButton.textContent;
        
        // Show loading state
        submitButton.disabled = true;
        submitButton.textContent = 'Saving...';
        
        fetch(form.action, {
            method: form.method,
            body: formData
        })
        .then(response => {
            if (response.ok) {
                showToast(successMessage || 'Changes saved successfully', 'success');
                return response.json();
            } else {
                throw new Error('Failed to save changes');
            }
        })
        .catch(error => {
            console.error('Error:', error);
            showToast('Failed to save changes', 'error');
        })
        .finally(() => {
            // Reset button state
            submitButton.disabled = false;
            submitButton.textContent = originalText;
        });
    });
}

// Auto-save functionality for settings forms
function enableAutoSave(formId, delay = 2000) {
    const form = document.getElementById(formId);
    if (!form) return;
    
    const inputs = form.querySelectorAll('input, select, textarea');
    let saveTimeout;
    
    inputs.forEach(input => {
        input.addEventListener('change', function() {
            clearTimeout(saveTimeout);
            saveTimeout = setTimeout(() => {
                const formData = new FormData(form);
                
                fetch(form.action, {
                    method: 'POST',
                    body: formData
                })
                .then(response => {
                    if (response.ok) {
                        showToast('Settings saved automatically', 'success');
                    }
                })
                .catch(error => {
                    console.error('Auto-save error:', error);
                });
            }, delay);
        });
    });
}

// Initialize components when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    // Close modals when clicking outside
    document.addEventListener('click', function(event) {
        const modals = document.querySelectorAll('[id$="-modal"]');
        modals.forEach(modal => {
            if (event.target === modal) {
                closeModal(modal.id);
            }
        });
    });
    
    // Handle escape key for modals
    document.addEventListener('keydown', function(event) {
        if (event.key === 'Escape') {
            const visibleModal = document.querySelector('[id$="-modal"]:not(.hidden)');
            if (visibleModal) {
                closeModal(visibleModal.id);
            }
        }
    });
    
    // Initialize delete user buttons
    initializeDeleteButtons();
    
    // Initialize search functionality
    initializeSearch();
    
    // Initialize tab functionality
    initializeTabs();
    
    // Initialize auto-save for settings forms
    enableAutoSave('profile-settings-form');
    enableAutoSave('notification-settings-form');
    enableAutoSave('privacy-settings-form');
    
    // Initialize form handlers
    handleFormSubmit('user-edit-form', 'Profile updated successfully');
    handleFormSubmit('password-change-form', 'Password updated successfully');
    
    // Initialize confirmation modal buttons
    initializeConfirmationModal();
});

// Initialize delete user buttons
function initializeDeleteButtons() {
    const deleteButtons = document.querySelectorAll('.delete-user-btn');
    deleteButtons.forEach(button => {
        button.addEventListener('click', function(e) {
            e.preventDefault();
            const userId = this.getAttribute('data-user-id');
            const userName = this.getAttribute('data-user-name');
            confirmDelete(userId, userName);
        });
    });
    
    // Initialize delete account button
    const deleteAccountBtn = document.getElementById('delete-account-btn');
    if (deleteAccountBtn) {
        deleteAccountBtn.addEventListener('click', function(e) {
            e.preventDefault();
            confirmDeleteAccount();
        });
    }
}

// Initialize search functionality
function initializeSearch() {
    const searchInput = document.querySelector('input[name="search"]');
    if (searchInput) {
        searchInput.addEventListener('input', function() {
            searchUsers(this.value);
        });
    }
}

// Initialize tab functionality
function initializeTabs() {
    const tabLinks = document.querySelectorAll('[role="tab"]');
    tabLinks.forEach(link => {
        link.addEventListener('click', function(e) {
            e.preventDefault();
            const targetTab = this.getAttribute('data-tab');
            switchTab(targetTab);
        });
    });
}

// Switch between tabs
function switchTab(tabId) {
    // Hide all tab content
    const tabContents = document.querySelectorAll('[role="tabpanel"]');
    tabContents.forEach(content => {
        content.classList.add('hidden');
    });
    
    // Remove active state from all tabs
    const tabLinks = document.querySelectorAll('[role="tab"]');
    tabLinks.forEach(link => {
        link.classList.remove('border-blue-500', 'text-blue-600');
        link.classList.add('border-transparent', 'text-gray-500');
    });
    
    // Show target tab content
    const targetContent = document.getElementById(`${tabId}-content`);
    if (targetContent) {
        targetContent.classList.remove('hidden');
    }
    
    // Activate target tab
    const targetTab = document.querySelector(`[data-tab="${tabId}"]`);
    if (targetTab) {
        targetTab.classList.remove('border-transparent', 'text-gray-500');
        targetTab.classList.add('border-blue-500', 'text-blue-600');
    }
}

// Initialize confirmation modal buttons
function initializeConfirmationModal() {
    const confirmBtn = document.getElementById('confirm-confirmation-modal');
    const cancelBtn = document.getElementById('cancel-confirmation-modal');
    
    if (confirmBtn) {
        confirmBtn.addEventListener('click', function() {
            confirmAction('confirmation-modal');
        });
    }
    
    if (cancelBtn) {
        cancelBtn.addEventListener('click', function() {
            closeModal('confirmation-modal');
        });
    }
}

// Utility functions
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

function formatDate(date) {
    return new Intl.DateTimeFormat('en-US', {
        year: 'numeric',
        month: 'long',
        day: 'numeric'
    }).format(new Date(date));
}

function formatDateTime(date) {
    return new Intl.DateTimeFormat('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    }).format(new Date(date));
}