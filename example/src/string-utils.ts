/**
 * String utility functions
 */

/**
 * Capitalizes the first letter of a string
 * @param str Input string
 * @returns String with first letter capitalized
 */
export function capitalize(str: string): string {
  if (str.length === 0) {
    return str;
  }
  return str.charAt(0).toUpperCase() + str.slice(1);
}

/**
 * Reverses a string
 * @param str Input string
 * @returns Reversed string
 */
export function reverse(str: string): string {
  return str.split('').reverse().join('');
}

/**
 * Checks if a string is a palindrome
 * @param str Input string
 * @returns True if palindrome, false otherwise
 */
export function isPalindrome(str: string): boolean {
  const cleaned = str.toLowerCase().replace(/\s/g, '');
  return cleaned === reverse(cleaned);
}

/**
 * Repeats a string n times
 * @param str Input string
 * @param times Number of times to repeat
 * @returns Repeated string
 */
export function repeat(str: string, times: number): string {
  if (times < 0) {
    throw new Error('Times must be non-negative');
  }
  return str.repeat(times);
}

/**
 * Truncates a string to a maximum length
 * @param str Input string
 * @param maxLength Maximum length
 * @param suffix Suffix to add if truncated (default: '...')
 * @returns Truncated string
 */
export function truncate(str: string, maxLength: number, suffix: string = '...'): string {
  if (str.length <= maxLength) {
    return str;
  }
  return str.slice(0, maxLength - suffix.length) + suffix;
}

