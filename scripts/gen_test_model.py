#!/usr/bin/env python3
"""Generate a tiny ONNX model (Y = X*W + B, 10→5 linear) for testing.

Requires: pip install onnx numpy

Usage:
    python scripts/gen_test_model.py

Output:
    stdlib/infer/testdata/tiny_linear.onnx
"""

import numpy as np
import onnx
from onnx import TensorProto, helper, numpy_helper


def main():
    # Weight matrix (10 x 5) and bias vector (5)
    W = numpy_helper.from_array(
        np.ones((10, 5), dtype=np.float32) * 0.1,
        name="W",
    )
    B = numpy_helper.from_array(
        np.zeros(5, dtype=np.float32),
        name="B",
    )

    # MatMul node: X @ W
    matmul = helper.make_node("MatMul", inputs=["input", "W"], outputs=["matmul_out"])

    # Add node: matmul_out + B
    add = helper.make_node("Add", inputs=["matmul_out", "B"], outputs=["output"])

    # Graph
    graph = helper.make_graph(
        nodes=[matmul, add],
        name="tiny_linear",
        inputs=[
            helper.make_tensor_value_info("input", TensorProto.FLOAT, [1, 10]),
        ],
        outputs=[
            helper.make_tensor_value_info("output", TensorProto.FLOAT, [1, 5]),
        ],
        initializer=[W, B],
    )

    model = helper.make_model(graph, opset_imports=[helper.make_opsetid("", 13)])
    model.ir_version = 7
    onnx.checker.check_model(model)

    path = "stdlib/infer/testdata/tiny_linear.onnx"
    onnx.save(model, path)
    print(f"Saved {path} ({len(open(path, 'rb').read())} bytes)")


if __name__ == "__main__":
    main()
