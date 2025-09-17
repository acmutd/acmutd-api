import csv
import openpyxl
import os
import csv

def excel_to_csv(excel_path, csv_path):
    ext = os.path.splitext(excel_path)[1].lower()
    rows = []
    if ext == '.xlsx':
        wb = openpyxl.load_workbook(excel_path)
        ws = wb["GradeDist"]
        ws["B1"] = "Catalog Nbr"
        for i, row in enumerate(ws.iter_rows(values_only=True)):
            row = list(row) if row else []
            if i == 0 and len(row) > 1:
                row[1] = "Catalog Nbr"
            if i > 0:
                for col in [22, 23, 24, 25, 26]:
                    if col < len(row) and (row[col] is None or row[col] == ""):
                        row[col] = ""
            rows.append(row)
    elif ext == '.xlsb':
        try:
            from pyxlsb import open_workbook
        except ImportError:
            raise ImportError("pyxlsb is required for .xlsb support. Install with 'pip install pyxlsb'.")
        with open_workbook(excel_path) as wb:
            ws = wb.get_sheet('GradeDist')
            for i, row in enumerate(ws.rows()):
                values = [cell.v if cell.v is not None else "" for cell in row]
                if i == 0 and len(values) > 1:
                    values[1] = "Catalog Nbr"
                if i > 0:
                    for col in [22, 23, 24, 25, 26]:
                        if col < len(values) and (values[col] is None or values[col] == ""):
                            values[col] = ""
                rows.append(values)
    else:
        raise ValueError("Unsupported file type: " + ext)

    min_cols = len(rows[0])
    padded_rows = []
    for row in rows:
        row = row + ["" for _ in range(min_cols - len(row))]
        padded_rows.append(row)

    with open(csv_path, "w", newline="") as f:
        writer = csv.writer(f, quoting=csv.QUOTE_MINIMAL)
        for row in padded_rows:
            writer.writerow(row)