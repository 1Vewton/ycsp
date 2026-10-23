"""晶体结构预测接口。"""

from http import HTTPStatus

from fastapi import APIRouter
from fastapi.responses import JSONResponse
from pydantic import BaseModel

from app import SERVICE_NAME
from app.constants.codes import CODE_PREDICTION_READY
from app.api.serialization import notice_to_dict, prediction_to_dict
from app.domain.notice import DomainError
from app.domain.prediction import predict_structure

router = APIRouter()


class PredictionRequest(BaseModel):
    """预测请求：只需要一个化学式。"""

    formula: str


@router.post("/predict")
def predict(request: PredictionRequest) -> JSONResponse:
    """按化学式预测晶体结构；输入有问题时返回可翻译的 code，不返回技术细节。"""
    try:
        result = predict_structure(request.formula)
    except DomainError as error:
        return JSONResponse(
            status_code=HTTPStatus.BAD_REQUEST,
            content={
                "service": SERVICE_NAME,
                "status": "error",
                **notice_to_dict(error.notice),
            },
        )

    return JSONResponse(
        status_code=HTTPStatus.OK,
        content={
            "service": SERVICE_NAME,
            "status": "ok",
            "code": CODE_PREDICTION_READY,
            "prediction": prediction_to_dict(result),
        },
    )
