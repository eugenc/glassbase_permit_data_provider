#!/usr/bin/env python3
"""Probe Miami Beach EnerGov search for date filtering (permits per week)."""
import json
import sys
import urllib.request

URL = "https://energovcss.miamibeachfl.gov/energovprod/selfservice/api/energov/search/search"
HEADERS = {
    "Content-Type": "application/json",
    "tenantId": "3",
    "Tyler-TenantUrl": "MiamiBeachFLProd",
    "Tyler-Tenant-Culture": "en-US",
}

# Minimal template matching migration shape; criteria blocks zeroed.
BASE = {
    "Keyword": "",
    "ExactMatch": False,
    "SearchModule": 1,
    "FilterModule": 2,
    "SearchMainAddress": False,
    "PlanCriteria": {
        "PageNumber": 0,
        "PageSize": 0,
        "ApplyDateFrom": None,
        "ApplyDateTo": None,
    },
    "PermitCriteria": {
        "PermitNumber": None,
        "PermitTypeId": None,
        "PermitWorkclassId": None,
        "PermitStatusId": None,
        "ProjectName": None,
        "IssueDateFrom": None,
        "IssueDateTo": None,
        "Address": None,
        "Description": None,
        "ExpireDateFrom": None,
        "ExpireDateTo": None,
        "FinalDateFrom": None,
        "FinalDateTo": None,
        "ApplyDateFrom": None,
        "ApplyDateTo": None,
        "SearchMainAddress": False,
        "ContactId": None,
        "TypeId": None,
        "WorkClassIds": None,
        "ParcelNumber": None,
        "ExcludeCases": None,
        "EnableDescriptionSearch": False,
        "PageNumber": 0,
        "PageSize": 0,
        "SortBy": None,
        "SortAscending": False,
    },
    "InspectionCriteria": {"PageNumber": 0, "PageSize": 0},
    "CodeCaseCriteria": {"PageNumber": 0, "PageSize": 0},
    "RequestCriteria": {"PageNumber": 0, "PageSize": 0},
    "BusinessLicenseCriteria": {"PageNumber": 0, "PageSize": 0},
    "ProfessionalLicenseCriteria": {"PageNumber": 0, "PageSize": 0},
    "LicenseCriteria": {"PageNumber": 0, "PageSize": 0},
    "ProjectCriteria": {"PageNumber": 0, "PageSize": 0},
    "ExcludeCases": None,
    "PageNumber": 1,
    "PageSize": 10,
    "SortBy": "relevance",
    "SortAscending": False,
}


def post(body: dict) -> dict:
    data = json.dumps(body).encode()
    req = urllib.request.Request(URL, data=data, headers=HEADERS, method="POST")
    with urllib.request.urlopen(req, timeout=60) as resp:
        return json.loads(resp.read().decode())


def count_summary(d: dict) -> str:
    if not d.get("Success"):
        msg = (d.get("ErrorMessage") or "")[:120]
        return f"Success=False err={msg!r}"
    r = d.get("Result") or {}
    pf = r.get("PermitsFound")
    tf = r.get("TotalFound")
    n = len(r.get("EntityResults") or [])
    return f"PermitsFound={pf} TotalFound={tf} firstPageRows={n}"


def main():
    week_apply = ("2020-01-01T00:00:00", "2020-01-07T23:59:59")
    week_apply_short = ("2020-01-01", "2020-01-07")

    tests = []

    # 1) Baseline — no permit dates
    b0 = json.loads(json.dumps(BASE))
    tests.append(("baseline (no dates)", b0))

    # 2) PermitCriteria ApplyDate ISO
    b = json.loads(json.dumps(BASE))
    b["PermitCriteria"]["ApplyDateFrom"] = week_apply[0]
    b["PermitCriteria"]["ApplyDateTo"] = week_apply[1]
    tests.append(("PermitCriteria ApplyDate ISO week", b))

    # 3) PermitCriteria ApplyDate short date
    b = json.loads(json.dumps(BASE))
    b["PermitCriteria"]["ApplyDateFrom"] = week_apply_short[0]
    b["PermitCriteria"]["ApplyDateTo"] = week_apply_short[1]
    tests.append(("PermitCriteria ApplyDate short week", b))

    # 4) PermitCriteria FinalDate week (often used in UI)
    b = json.loads(json.dumps(BASE))
    b["PermitCriteria"]["FinalDateFrom"] = week_apply[0]
    b["PermitCriteria"]["FinalDateTo"] = week_apply[1]
    tests.append(("PermitCriteria FinalDate ISO week", b))

    # 5) PermitCriteria IssueDate week
    b = json.loads(json.dumps(BASE))
    b["PermitCriteria"]["IssueDateFrom"] = week_apply[0]
    b["PermitCriteria"]["IssueDateTo"] = week_apply[1]
    tests.append(("PermitCriteria IssueDate ISO week", b))

    # 6) SearchModule 2 (permit?) — may 500
    b = json.loads(json.dumps(BASE))
    b["SearchModule"] = 2
    b["FilterModule"] = None
    b["PermitCriteria"]["ApplyDateFrom"] = week_apply[0]
    b["PermitCriteria"]["ApplyDateTo"] = week_apply[1]
    tests.append(("SearchModule=2 + ApplyDate week", b))

    # 7) Root-level ApplyDateFrom/To (nonstandard)
    b = json.loads(json.dumps(BASE))
    b["ApplyDateFrom"] = week_apply[0]
    b["ApplyDateTo"] = week_apply[1]
    tests.append(("root ApplyDateFrom/To week", b))

    # 8) SortBy apply date + PermitCriteria ApplyDate
    b = json.loads(json.dumps(BASE))
    b["SortBy"] = "ApplyDate"
    b["SortAscending"] = True
    b["PermitCriteria"]["ApplyDateFrom"] = week_apply[0]
    b["PermitCriteria"]["ApplyDateTo"] = week_apply[1]
    tests.append(("SortBy=ApplyDate + PermitCriteria Apply week", b))

    # 9) PlanCriteria ApplyDate only (wild)
    b = json.loads(json.dumps(BASE))
    b["PlanCriteria"]["ApplyDateFrom"] = week_apply[0]
    b["PlanCriteria"]["ApplyDateTo"] = week_apply[1]
    tests.append(("PlanCriteria ApplyDate week only", b))

    rows = []
    for name, body in tests:
        try:
            d = post(body)
            rows.append((name, count_summary(d)))
        except Exception as e:
            rows.append((name, f"REQUEST_FAILED {e}"))

    # Pick winner: smallest PermitsFound if Success
    print("--- Miami Beach EnerGov /search/search — week 2020-01-01 .. 2020-01-07 ---\n")
    for name, summary in rows:
        print(f"{name}\n  -> {summary}\n")

    # If any test shows filter working, sample first record ApplyDate
    for name, body in tests:
        try:
            b = json.loads(json.dumps(body))
            b["PageNumber"] = 1
            b["PageSize"] = 3
            d = post(b)
            if not d.get("Success"):
                continue
            ents = (d.get("Result") or {}).get("EntityResults") or []
            if not ents:
                continue
            dates = [e.get("ApplyDate") for e in ents]
            print(f"Sample ApplyDate from {name!r}: {dates}")
        except Exception:
            pass


if __name__ == "__main__":
    main()
