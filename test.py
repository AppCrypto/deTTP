# pyecc测试secp256k1指数运算的时间开销
# import time
# from py_ecc.secp256k1 import secp256k1 as secp256k1
# from charm.toolbox.pairinggroup import PairingGroup, G1, G2, ZR,pair
# from charm.toolbox.ecgroup import ECGroup,G,ZR
# from umbral.curve_point import CurvePoint
# from umbral.curve_scalar import CurveScalar
# from umbral import SecretKey
# from py_ecc.bn128 import G1 as BNG1
# from py_ecc.bn128 import G2 as BNG2
# from py_ecc.bn128 import add, multiply, neg, pairing, is_on_curve
# from py_ecc.bn128 import curve_order as CURVE_ORDER
# import random

# def test_pyecc_secp256k1(num_tests):
#     # 使用基点G进行指数运算
#     g = secp256k1.G
#     scalar = 108597297017729806015314184954651673007004806959293110077020910939277895371122  # 随意选择一个标量

#     # 记录所有指数运算的时间开销
#     total_time = 0
#     start_time=time.time()
#     for _ in range(num_tests):
#         # 执行标量乘法（指数运算）
#         result = secp256k1.multiply(g, scalar)

#         # 计算单次指数运算的时间并累加到总时间
#     total_time = (time.time() - start_time)

#     # 计算平均时间开销
#     average_time = total_time / num_tests

#     print(f"py_ecc_secp256k1指数运算平均时间: {average_time:.6f}秒")

# def test_pyecc_bn128(num_tests):
#     # 使用基点G进行指数运算
    
#     scalar = random.randint(0,CURVE_ORDER)

#     # 记录所有指数运算的时间开销
#     total_time = 0
#     start_time=time.time()
#     for _ in range(num_tests):
#         x=multiply(BNG1,scalar) 
#     total_time = (time.time() - start_time)
#     average_time = total_time / num_tests

    
#     total_time2 = 0
#     start_time=time.time()
#     for _ in range(num_tests):
#         x=multiply(BNG2,scalar) 
#     total_time2 = (time.time() - start_time)
#     average_time2 = total_time2 / num_tests
    
#     # BNGT = pairing(BNG2,BNG1)

#     # total_timeT = 0
#     # start_time=time.time()
#     # for _ in range(num_tests):
#     #     x=multiply(BNGT,scalar) 
#     # total_timeT = (time.time() - start_time)
#     # average_timeT = total_timeT / num_tests



#     print(f"py_ecc_bn128 G1指数运算平均时间: {average_time:.6f}秒")
#     print(f"py_ecc_bn128 G2指数运算平均时间: {average_time2:.6f}秒")
#     # print(f"py_ecc_bn128 GT指数运算平均时间: {average_timeT:.6f}秒")

# def test_charm_crypto_secp256k1(num_tests):
#     # 初始化一个椭圆曲线群
#     group = ECGroup(714)

#     # 随机生成
#     point = group.random(G)
#     a = group.random(ZR)
#     # print(a)
#     total_time = 0
#     start_time = time.time()
#     for _ in range(num_tests):
#         new_point = point ** a
        
#     total_time = (time.time() - start_time)
#     average_time = total_time / num_tests
#     print(f"charm_secp256k1指数运算平均时间: {average_time:.6f}秒")

# def test_charm_crypto_BN254(num_tests):
#     # 初始化一个PairingGroup
#     group = PairingGroup('BN254')

#     # 随机选择一个元素和一个标量
#     g1 = group.random(G1)
#     g2 = group.random(G2)
#     gt = group.pair_prod(g1, g2)
#     scalar = group.random(ZR)

#     # 测试指数运算的时间开销
#     total_exp_time_G1 = 0
#     start_time = time.time()
#     for _ in range(num_tests):        
#         result = g1 ** scalar
#     total_exp_time_G1 = (time.time() - start_time)
#     average_exp_time_G1 = total_exp_time_G1 / num_tests

#     # 测试指数运算的时间开销
#     total_exp_time_G2 = 0
#     start_time = time.time()
#     for _ in range(num_tests):
#         result = g2 ** scalar        
#     total_exp_time_G2 = (time.time() - start_time)
#     average_exp_time_G2 = total_exp_time_G2 / num_tests

#     # 测试指数运算的时间开销
#     total_exp_time_GT = 0    
#     start_time = time.time()
#     for _ in range(num_tests):        
#         result = gt ** scalar
#     total_exp_time_GT = (time.time() - start_time)
#     average_exp_time_GT = total_exp_time_GT / num_tests

#     # 测试双线性对运算的时间开销
#     total_pairing_time = 0
#     start_time = time.time()
#     for _ in range(num_tests):
#         # result = pair(g1, g2)
#         result = group.pair_prod(g1, g2)
#         # print(result)
#     total_pairing_time = (time.time() - start_time)
#     average_pairing_time = total_pairing_time / num_tests

#     print(f"charm-crypto_BN254_G1指数运算平均时间: {average_exp_time_G1:.6f}秒")
#     print(f"charm-crypto_BN254_G2指数运算平均时间: {average_exp_time_G2:.6f}秒")
#     print(f"charm-crypto_BN254_GT指数运算平均时间: {average_exp_time_GT:.6f}秒")
#     print(f"charm-crypto_BN254双线性对运算平均时间: {average_pairing_time:.6f}秒")

# def test_pyUmbral_secp256k1(num_tests):

#     q = CurveScalar.random_nonzero()
#     # print(q.__attr__)
#     q.from_int(108597297017729806015314184954651673007004806959293110077020910939277895371122)
#     g = CurvePoint.generator()
#     # 测试指数运算的时间开销
#     total_exp_time = 0
#     start_time = time.time()
#     for _ in range(num_tests):        
#         X_q = g*q  # 标量乘法
#         # delegating_sk = SecretKey.random()
#         # delegating_pk = delegating_sk.public_key()
        
#     total_exp_time = (time.time() - start_time)
#     average_exp_time = total_exp_time / num_tests

#     print(f"pyUmbral指数运算平均时间: {average_exp_time:.6f}秒")


# test_pyecc_bn128(10)
# test_pyUmbral_secp256k1(100)
# test_charm_crypto_secp256k1(100)
# test_charm_crypto_BN254(100)


from py_ecc.fields import (
    bn128_FQ as FQ,
    bn128_FQ2 as FQ2,
    bn128_FQ12 as FQ12,
    bn128_FQP as FQP,
)
from py_ecc.typing import (
    Field,
    GeneralPoint,
    Point2D,
)
# Check if a point is the point at infinity
def is_inf(pt: GeneralPoint[Field]) -> bool:
    return pt is None



G1 = (FQ(int('030644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd3', 16)), FQ(int("15ed738c0e0a7c92e7845f96b2ae9c0a68a6a449e3538fc7ff3ebf7a5a18a2c4",16)))
def is_on_curve(pt: Point2D[Field], b: Field) -> bool:
    if is_inf(pt):
        return True
    x, y = pt
    return y**2 - x**3 == b

b = FQ(3)
print(is_on_curve(G1, b))
# print(multiply(G1, 21888242871839275222246405745257275088548364400416034343698204186575808495617))
# assert is_on_curve(G2, b2)



