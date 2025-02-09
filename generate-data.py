import csv
import random
import datetime

def generate_random_date(start_year=2023, end_year=2025):
    """Generate a random date string between start_year and end_year (YYYY-MM-DD)."""
    start_date = datetime.date(start_year, 1, 1)
    end_date = datetime.date(end_year, 12, 31)
    delta_days = (end_date - start_date).days

    random_days = random.randint(0, delta_days)
    random_date = start_date + datetime.timedelta(days=random_days)
    return random_date.strftime("%Y-%m-%d")

def generate_disclosure_csv(filename, rows=1000000):
    """
    Generates a CSV with columns:
      company_id,date,dis_1,dis_2,dis_3,dis_4
    Each row is randomly generated, with float values for dis_x fields.
    """
    headers = ["company_id", "date", "dis_1", "dis_2", "dis_3", "dis_4"]
    with open(filename, "w", newline="") as f:
        writer = csv.writer(f)
        writer.writerow(headers)
        for _ in range(rows):
            company_id = random.randint(1000, 1100)
            date_year_only = random.randint(2023, 2025)
            # random floats for disclosures
            dis_1 = round(random.uniform(0, 100), 2)
            dis_2 = round(random.uniform(0, 100), 2)
            dis_3 = round(random.uniform(0, 100), 2)
            dis_4 = round(random.uniform(0, 100), 2)
            writer.writerow([company_id, date_year_only, dis_1, dis_2, dis_3, dis_4])

def generate_emissions_csv(filename, rows=1000000):
    """
    Generates a CSV with columns:
      company_id,date,emi_1,emi_2,emi_3,emi_4
    """
    headers = ["company_id", "date", "emi_1", "emi_2", "emi_3", "emi_4"]
    with open(filename, "w", newline="") as f:
        writer = csv.writer(f)
        writer.writerow(headers)
        for _ in range(rows):
            company_id = random.randint(1000, 1100)
            date_str = generate_random_date()  # full date
            # Random chance of missing value (None) for demonstration
            emi_1 = round(random.uniform(0, 100), 2) if random.random() > 0.05 else ""  
            emi_2 = round(random.uniform(0, 100), 2)
            emi_3 = round(random.uniform(0, 100), 2)
            emi_4 = round(random.uniform(0, 100), 2)
            writer.writerow([company_id, date_str, emi_1, emi_2, emi_3, emi_4])

def generate_waste_csv(filename, rows=1000000):
    """
    Generates a CSV with columns:
      company_id,date,was_1,was_2,was_3,was_4
    """
    headers = ["company_id", "date", "was_1", "was_2", "was_3", "was_4"]
    with open(filename, "w", newline="") as f:
        writer = csv.writer(f)
        writer.writerow(headers)
        for _ in range(rows):
            company_id = random.randint(1000, 1100)
            date_str = generate_random_date()
            was_1 = round(random.uniform(0, 100), 2)
            was_2 = round(random.uniform(0, 100), 2)
            was_3 = round(random.uniform(0, 100), 2)
            was_4 = round(random.uniform(0, 100), 2)
            writer.writerow([company_id, date_str, was_1, was_2, was_3, was_4])

if __name__ == "__main__":
    # Generate the 3 CSVs with 10,000 rows each (customize row count if needed):
    generate_disclosure_csv("disclosure_10k.csv", rows=1000000)
    generate_emissions_csv("emissions_10k.csv", rows=1000000)
    generate_waste_csv("waste_10k.csv", rows=1000000)
    
    print("Finished generating 3 CSV files with 10k rows each!")
