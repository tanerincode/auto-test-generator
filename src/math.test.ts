import { add, subtract, multiply, divide, average } from './math';

describe('Math Utility Functions', () => {
  describe('add', () => {
    it('should add two positive numbers correctly', () => {
      expect(add(2, 3)).toBe(5);
      expect(add(10, 15)).toBe(25);
    });

    it('should add negative numbers correctly', () => {
      expect(add(-2, -3)).toBe(-5);
      expect(add(-10, 5)).toBe(-5);
      expect(add(10, -5)).toBe(5);
    });

    it('should handle zero values', () => {
      expect(add(0, 0)).toBe(0);
      expect(add(5, 0)).toBe(5);
      expect(add(0, 5)).toBe(5);
    });

    it('should handle decimal numbers', () => {
      expect(add(1.5, 2.5)).toBe(4);
      expect(add(0.1, 0.2)).toBeCloseTo(0.3);
    });

    it('should handle very large numbers', () => {
      expect(add(Number.MAX_SAFE_INTEGER, 0)).toBe(Number.MAX_SAFE_INTEGER);
      expect(add(1e10, 2e10)).toBe(3e10);
    });

    it('should handle Infinity', () => {
      expect(add(Infinity, 5)).toBe(Infinity);
      expect(add(-Infinity, 5)).toBe(-Infinity);
      expect(add(Infinity, -Infinity)).toBeNaN();
    });
  });

  describe('subtract', () => {
    it('should subtract two positive numbers correctly', () => {
      expect(subtract(5, 3)).toBe(2);
      expect(subtract(10, 4)).toBe(6);
    });

    it('should handle negative results', () => {
      expect(subtract(3, 5)).toBe(-2);
      expect(subtract(-5, -3)).toBe(-2);
    });

    it('should handle zero values', () => {
      expect(subtract(0, 0)).toBe(0);
      expect(subtract(5, 0)).toBe(5);
      expect(subtract(0, 5)).toBe(-5);
    });

    it('should handle decimal numbers', () => {
      expect(subtract(2.5, 1.5)).toBe(1);
      expect(subtract(0.3, 0.1)).toBeCloseTo(0.2);
    });

    it('should handle negative numbers', () => {
      expect(subtract(-5, -3)).toBe(-2);
      expect(subtract(-5, 3)).toBe(-8);
      expect(subtract(5, -3)).toBe(8);
    });

    it('should handle Infinity', () => {
      expect(subtract(Infinity, 5)).toBe(Infinity);
      expect(subtract(-Infinity, 5)).toBe(-Infinity);
      expect(subtract(Infinity, Infinity)).toBeNaN();
    });
  });

  describe('multiply', () => {
    it('should multiply two positive numbers correctly', () => {
      expect(multiply(2, 3)).toBe(6);
      expect(multiply(4, 5)).toBe(20);
    });

    it('should handle zero multiplication', () => {
      expect(multiply(0, 5)).toBe(0);
      expect(multiply(5, 0)).toBe(0);
      expect(multiply(0, 0)).toBe(0);
    });

    it('should handle negative numbers', () => {
      expect(multiply(-2, 3)).toBe(-6);
      expect(multiply(2, -3)).toBe(-6);
      expect(multiply(-2, -3)).toBe(6);
    });

    it('should handle decimal numbers', () => {
      expect(multiply(2.5, 4)).toBe(10);
      expect(multiply(0.1, 0.2)).toBeCloseTo(0.02);
    });

    it('should handle multiplication by one', () => {
      expect(multiply(5, 1)).toBe(5);
      expect(multiply(1, 5)).toBe(5);
      expect(multiply(-5, 1)).toBe(-5);
    });

    it('should handle Infinity', () => {
      expect(multiply(Infinity, 5)).toBe(Infinity);
      expect(multiply(-Infinity, 5)).toBe(-Infinity);
      expect(multiply(Infinity, 0)).toBeNaN();
    });
  });

  describe('divide', () => {
    it('should divide two positive numbers correctly', () => {
      expect(divide(6, 2)).toBe(3);
      expect(divide(15, 3)).toBe(5);
    });

    it('should handle decimal division', () => {
      expect(divide(5, 2)).toBe(2.5);
      expect(divide(1, 3)).toBeCloseTo(0.3333333333333333);
    });

    it('should handle negative numbers', () => {
      expect(divide(-6, 2)).toBe(-3);
      expect(divide(6, -2)).toBe(-3);
      expect(divide(-6, -2)).toBe(3);
    });

    it('should handle division by one', () => {
      expect(divide(5, 1)).toBe(5);
      expect(divide(-5, 1)).toBe(-5);
    });

    it('should handle zero dividend', () => {
      expect(divide(0, 5)).toBe(0);
      expect(divide(0, -5)).toBe(-0);
    });

    it('should throw error when dividing by zero', () => {
      expect(() => divide(5, 0)).toThrow('Division by zero');
      expect(() => divide(-5, 0)).toThrow('Division by zero');
      expect(() => divide(0, 0)).toThrow('Division by zero');
    });

    it('should handle Infinity', () => {
      expect(divide(Infinity, 5)).toBe(Infinity);
      expect(divide(-Infinity, 5)).toBe(-Infinity);
      expect(divide(5, Infinity)).toBe(0);
    });

    // Edge case: very small divisor (close to zero but not exactly zero)
    it('should handle very small divisors', () => {
      expect(divide(1, 1e-10)).toBe(1e10);
      expect(divide(1, -1e-10)).toBe(-1e10);
    });
  });

  describe('average', () => {
    it('should calculate average of positive numbers', () => {
      expect(average([1, 2, 3, 4, 5])).toBe(3);
      expect(average([10, 20, 30])).toBe(20);
    });

    it('should handle single number array', () => {
      expect(average([5])).toBe(5);
      expect(average([-5])).toBe(-5);
    });

    it('should handle empty array', () => {
      expect(average([])).toBe(0);
    });

    it('should handle negative numbers', () => {
      expect(average([-1, -2, -3])).toBe(-2);
      expect(average([-5, 5])).toBe(0);
    });

    it('should handle decimal numbers', () => {
      expect(average([1.5, 2.5, 3.5])).toBeCloseTo(2.5);
      expect(average([0.1, 0.2, 0.3])).toBeCloseTo(0.2);
    });

    it('should handle zero values', () => {
      expect(average([0, 0, 0])).toBe(0);
      expect(average([0, 5, 10])).toBe(5);
    });

    it('should handle mixed positive and negative numbers', () => {
      expect(average([-10, 0, 10])).toBe(0);
      expect(average([-5, -3, 2, 6])).toBe(0);
    });

    it('should handle large arrays', () => {
      const largeArray = Array.from({ length: 1000 }, (_, i) => i + 1);
      expect(average(largeArray)).toBe(500.5);
    });

    it('should handle arrays with Infinity', () => {
      expect(average([Infinity, 1, 2])).toBe(Infinity);
      expect(average([-Infinity, 1, 2])).toBe(-Infinity);
      expect(average([Infinity, -Infinity])).toBeNaN();
    });

    // Edge case: very large numbers that might cause overflow
    it('should handle very large numbers', () => {
      const largeNumbers = [Number.MAX_SAFE_INTEGER, Number.MAX_SAFE_INTEGER];
      expect(average(largeNumbers)).toBe(Number.MAX_SAFE_INTEGER);
    });

    // Edge case: very small numbers
    it('should handle very small numbers', () => {
      expect(average([1e-10, 2e-10, 3e-10])).toBeCloseTo(2e-10);
    });
  });

  // Integration tests - testing functions together
  describe('Integration Tests', () => {
    it('should work correctly when chaining operations', () => {
      const result1 = add(5, 3); // 8
      const result2 = multiply(result1, 2); // 16
      const result3 = divide(result2, 4); // 4
      const result4 = subtract(result3, 1); // 3
      expect(result4).toBe(3);
    });

    it('should calculate average of computed values', () => {
      const values = [
        add(1, 2), // 3
        subtract(10, 5), // 5
        multiply(2, 3), // 6
        divide(20, 4) // 5
      ];
      expect(average(values)).toBeCloseTo(4.75);
    });
  });

  // Type safety tests (TypeScript specific)
  describe('Type Safety', () => {
    it('should handle NaN inputs gracefully', () => {
      expect(add(NaN, 5)).toBeNaN();
      expect(subtract(NaN, 5)).toBeNaN();
      expect(multiply(NaN, 5)).toBeNaN();
      expect(divide(NaN, 5)).toBeNaN();
      expect(average([NaN, 1, 2])).toBeNaN();
    });

    it('should handle undefined and null as NaN (JavaScript coercion)', () => {
      // Note: TypeScript would prevent these at compile time, but testing runtime behavior
      expect(add(undefined as any, 5)).toBeNaN();
      expect(add(null as any, 5)).toBe(5); // null coerces to 0
      expect(average([undefined as any, 1, 2])).toBeNaN();
      expect(average([null as any, 1, 2])).toBeCloseTo(1); // null coerces to 0
    });
  });
});
