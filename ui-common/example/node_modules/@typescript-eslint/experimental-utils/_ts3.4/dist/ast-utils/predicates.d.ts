import { AST_NODE_TYPES, TSESTree } from '../ts-estree';
declare function isOptionalChainPunctuator(token: TSESTree.Token): token is TSESTree.PunctuatorToken & {
    value: '?.';
};
declare function isNotOptionalChainPunctuator(token: TSESTree.Token): token is Exclude<TSESTree.Token, TSESTree.PunctuatorToken & {
    value: '?.';
}>;
declare function isNonNullAssertionPunctuator(token: TSESTree.Token): token is TSESTree.PunctuatorToken & {
    value: '!';
};
declare function isNotNonNullAssertionPunctuator(token: TSESTree.Token): token is Exclude<TSESTree.Token, TSESTree.PunctuatorToken & {
    value: '!';
}>;
/**
 * Returns true if and only if the node represents: foo?.() or foo.bar?.()
 */
declare const isOptionalCallExpression: (node: TSESTree.Node | null | undefined) => node is TSESTree.CallExpression & {
    type: AST_NODE_TYPES.CallExpression;
} & {
    optional: true;
};
/**
 * Returns true if and only if the node represents logical OR
 */
declare const isLogicalOrOperator: (node: TSESTree.Node | null | undefined) => node is TSESTree.LogicalExpression & {
    type: AST_NODE_TYPES.LogicalExpression;
} & {
    operator: "||";
};
/**
 * Checks if a node is a type assertion:
 * ```
 * x as foo
 * <foo>x
 * ```
 */
declare function isTypeAssertion(node: TSESTree.Node | undefined | null): node is TSESTree.TSAsExpression | TSESTree.TSTypeAssertion;
declare const isVariableDeclarator: (node: TSESTree.Node | null | undefined) => node is TSESTree.VariableDeclarator & {
    type: AST_NODE_TYPES.VariableDeclarator;
};
declare function isFunction(node: TSESTree.Node | undefined): node is TSESTree.ArrowFunctionExpression | TSESTree.FunctionDeclaration | TSESTree.FunctionExpression;
declare function isFunctionType(node: TSESTree.Node | undefined): node is TSESTree.TSCallSignatureDeclaration | TSESTree.TSConstructorType | TSESTree.TSConstructSignatureDeclaration | TSESTree.TSEmptyBodyFunctionExpression | TSESTree.TSFunctionType | TSESTree.TSMethodSignature;
declare function isFunctionOrFunctionType(node: TSESTree.Node | undefined): node is TSESTree.ArrowFunctionExpression | TSESTree.FunctionDeclaration | TSESTree.FunctionExpression | TSESTree.TSCallSignatureDeclaration | TSESTree.TSConstructorType | TSESTree.TSConstructSignatureDeclaration | TSESTree.TSEmptyBodyFunctionExpression | TSESTree.TSFunctionType | TSESTree.TSMethodSignature;
declare const isTSFunctionType: (node: TSESTree.Node | null | undefined) => node is TSESTree.TSFunctionType & {
    type: AST_NODE_TYPES.TSFunctionType;
};
declare const isTSConstructorType: (node: TSESTree.Node | null | undefined) => node is TSESTree.TSConstructorType & {
    type: AST_NODE_TYPES.TSConstructorType;
};
declare function isClassOrTypeElement(node: TSESTree.Node | undefined): node is TSESTree.ClassElement | TSESTree.TypeElement;
/**
 * Checks if a node is a constructor method.
 */
declare const isConstructor: (node: TSESTree.Node | null | undefined) => node is (TSESTree.MethodDefinitionComputedName & {
    type: AST_NODE_TYPES.MethodDefinition;
} & {
    kind: "constructor";
}) | (TSESTree.MethodDefinitionNonComputedName & {
    type: AST_NODE_TYPES.MethodDefinition;
} & {
    kind: "constructor";
});
/**
 * Checks if a node is a setter method.
 */
declare function isSetter(node: TSESTree.Node | undefined): node is TSESTree.MethodDefinition | TSESTree.Property;
declare const isIdentifier: (node: TSESTree.Node | null | undefined) => node is TSESTree.Identifier & {
    type: AST_NODE_TYPES.Identifier;
};
/**
 * Checks if a node represents an `await …` expression.
 */
declare const isAwaitExpression: (node: TSESTree.Node | null | undefined) => node is TSESTree.AwaitExpression & {
    type: AST_NODE_TYPES.AwaitExpression;
};
/**
 * Checks if a possible token is the `await` keyword.
 */
declare function isAwaitKeyword(node: TSESTree.Token | undefined | null): node is TSESTree.IdentifierToken & {
    value: 'await';
};
declare function isLoop(node: TSESTree.Node | undefined | null): node is TSESTree.DoWhileStatement | TSESTree.ForStatement | TSESTree.ForInStatement | TSESTree.ForOfStatement | TSESTree.WhileStatement;
export { isAwaitExpression, isAwaitKeyword, isConstructor, isClassOrTypeElement, isFunction, isFunctionOrFunctionType, isFunctionType, isIdentifier, isLoop, isLogicalOrOperator, isNonNullAssertionPunctuator, isNotNonNullAssertionPunctuator, isNotOptionalChainPunctuator, isOptionalChainPunctuator, isOptionalCallExpression, isSetter, isTSConstructorType, isTSFunctionType, isTypeAssertion, isVariableDeclarator, };
//# sourceMappingURL=predicates.d.ts.map
