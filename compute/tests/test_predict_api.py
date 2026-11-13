"""预测接口的应答契约：成功给结构数据，失败给可翻译的 code。"""

from fastapi.testclient import TestClient

from app.constants.codes import CODE_PREDICTION_READY, CODE_REQUEST_INVALID_BODY
from app.constants.codes import ERROR_FORMULA_CHARGE_UNBALANCED
from app.constants.structures import PROTOTYPE_ROCK_SALT, PROTOTYPE_RUTILE
from app.main import app

client = TestClient(app)


def test_predict_returns_cell_and_atoms():
    response = client.post("/predict", json={"formula": "NaCl"})

    assert response.status_code == 200
    payload = response.json()
    assert payload["status"] == "ok"
    assert payload["code"] == CODE_PREDICTION_READY

    candidate = payload["prediction"]["candidates"][0]
    assert candidate["prototype"] == PROTOTYPE_ROCK_SALT
    assert candidate["cell"]["edge_a_angstrom"] == 5.66
    assert len(candidate["atoms"]) == 8
    assert candidate["atoms"][0]["symbol"] == "Na"


def test_predict_metal_family_payload():
    """金属这一支走的是实验结构表，应答里要给前端同样的结构数据。"""
    response = client.post("/predict", json={"formula": "Cu"})

    assert response.status_code == 200
    prediction = response.json()["prediction"]
    assert prediction["family"] == "metallic"

    candidate = prediction["candidates"][0]
    assert candidate["coordination_number"] == 12
    assert {atom["symbol"] for atom in candidate["atoms"]} == {"Cu"}


def test_predict_rutile_payload_has_tetragonal_cell():
    """金红石型是四方晶胞：a = b ≠ c，前端要拿到完整的 6 个晶胞参数。"""
    response = client.post("/predict", json={"formula": "TiO2"})

    assert response.status_code == 200
    candidate = response.json()["prediction"]["candidates"][0]
    assert candidate["prototype"] == PROTOTYPE_RUTILE
    assert candidate["coordination_number"] == 6
    assert candidate["cell"]["edge_a_angstrom"] == candidate["cell"]["edge_b_angstrom"]
    assert candidate["cell"]["edge_c_angstrom"] != candidate["cell"]["edge_a_angstrom"]
    assert len(candidate["atoms"]) == 6


def test_predict_error_is_translatable():
    response = client.post("/predict", json={"formula": "NaCl2"})

    assert response.status_code == 400
    payload = response.json()
    assert payload["status"] == "error"
    assert payload["code"] == ERROR_FORMULA_CHARGE_UNBALANCED
    assert payload["params"]["formula"] == "NaCl2"


def test_predict_invalid_body_uses_same_error_shape():
    """缺 formula 字段时也不能把框架的英文报错直接甩给界面。"""
    response = client.post("/predict", json={})

    assert response.status_code == 400
    payload = response.json()
    assert payload["status"] == "error"
    assert payload["code"] == CODE_REQUEST_INVALID_BODY
