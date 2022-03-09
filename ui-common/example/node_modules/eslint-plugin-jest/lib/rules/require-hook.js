"use strict";

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.default = void 0;

var _experimentalUtils = require("@typescript-eslint/experimental-utils");

var _utils = require("./utils");

const isJestFnCall = node => {
  var _getNodeName;

  if ((0, _utils.isDescribeCall)(node) || (0, _utils.isTestCaseCall)(node) || (0, _utils.isHook)(node)) {
    return true;
  }

  return !!((_getNodeName = (0, _utils.getNodeName)(node)) !== null && _getNodeName !== void 0 && _getNodeName.startsWith('jest.'));
};

const shouldBeInHook = node => {
  switch (node.type) {
    case _experimentalUtils.AST_NODE_TYPES.ExpressionStatement:
      return shouldBeInHook(node.expression);

    case _experimentalUtils.AST_NODE_TYPES.CallExpression:
      return !isJestFnCall(node);

    default:
      return false;
  }
};

var _default = (0, _utils.createRule)({
  name: __filename,
  meta: {
    docs: {
      category: 'Best Practices',
      description: 'Require setup and teardown code to be within a hook',
      recommended: false
    },
    messages: {
      useHook: 'This should be done within a hook'
    },
    type: 'suggestion',
    schema: []
  },
  defaultOptions: [],

  create(context) {
    const checkBlockBody = body => {
      for (const statement of body) {
        if (shouldBeInHook(statement)) {
          context.report({
            node: statement,
            messageId: 'useHook'
          });
        }
      }
    };

    return {
      Program(program) {
        checkBlockBody(program.body);
      },

      CallExpression(node) {
        if (!(0, _utils.isDescribeCall)(node) || node.arguments.length < 2) {
          return;
        }

        const [, testFn] = node.arguments;

        if (!(0, _utils.isFunction)(testFn) || testFn.body.type !== _experimentalUtils.AST_NODE_TYPES.BlockStatement) {
          return;
        }

        checkBlockBody(testFn.body.body);
      }

    };
  }

});

exports.default = _default;