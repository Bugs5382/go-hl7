# 🧬 Segment Reference

> A quick-lookup matrix for every HL7 v2.x segment supported by go-hl7's typed builders, plus links to the canonical [Caristix](https://hl7-definition.caristix.com/v2/) field reference.

Every segment is exposed by `*hl7.HL7_BASE` as a public `Build<SEGNAME>(props)` method. Version-specific constructors (`NewHL7_2_1` … `NewHL7_2_8`) configure the segments that exist in their version. Calling a `Build<SEGNAME>` for a segment that doesn't exist in the active version raises `HL7ValidationError("Segment <NAME> is not part of HL7 v<X>")` at runtime — sourced from the per-segment `SegmentSpec` (see below).

## 🧾 Table of Contents

1. [How to read the matrix](#-how-to-read-the-matrix)
2. [SegmentSpec catalogue (all 187 segments)](#-segmentspec-catalogue-all-187-segments)
3. [HL7 value tables (version-aware enforcement)](#-hl7-value-tables-version-aware-enforcement)
4. [Always-available (base) segments](#-always-available-base-segments)
5. [Compatibility matrix](#-compatibility-matrix)
6. [By category](#-by-category)
7. [Per-segment cheat-sheet](#-per-segment-cheat-sheet)

---

## 🗺️ How to read the matrix

- ✅ — segment is supported in this version's builder.
- ➖ — segment was not yet defined in HL7 at that version (or simply isn't implemented).
- 🔁 — same segment was extended in this version (more fields). The library uses the most recent definition automatically when you instantiate that version's constructor.
- All segments are inherited downstream — once a segment first appears in version V, every later version (`NewHL7_2_(V+1)`, … `NewHL7_2_8`) also supports it.

> 💡 **Tip:** Always start the message with `BuildMSH(...)`. Calling any other `Build*` first panics with `HL7FatalError("MSH Header must be built first.")`.

---

## 📚 SegmentSpec catalogue (all 187 segments)

Beyond the typed builders below, the library ships a complete **machine-readable catalogue** of every HL7 v2 segment across versions 2.1 → 2.8 — generated from the [Caristix HL7 Definition API](https://hl7-definition.caristix.com/v2/) and committed to the repo (no runtime network calls). It lives in the `client/hl7/metadata` package.

```go
import "github.com/Bugs5382/go-hl7/client/hl7/metadata"

// Every spec carries per-version field-level usage codes (R/O/B/W/D/X).
ecd := metadata.SEGMENT_SPECS["ECD"]
ecd.Versions                          // [2.4 2.5 … 2.8]
ecd.Fields[3].Usage["2.8"]            // "W" — ECD.4 was withdrawn in 2.8

// Composite fields (XAD, XPN, CE, CWE, …) carry sub-component metadata too.
var pid11 metadata.FieldSpec
for _, f := range metadata.SEGMENT_SPECS["PID"].Fields {
    if f.Num == 11 {
        pid11 = f
        break
    }
}
names := make([]string, 0, len(pid11.Components))
for _, c := range pid11.Components {
    names = append(names, c.Name)
}
// names → ["Street Address", "Other Designation", "City",
//          "State Or Province", "Zip Or Postal Code", "Country", …23 total]
```

Use this with `builder.BuildSegment(name, props)` to cover the long tail of segments that don't have a hand-tuned typed method:

```go
builder.
    BuildMSH(hl7.Props{"msh_9": "ADT^A01", "msh_10": "X", "msh_11": "P"}).
    BuildSegment("ABS", hl7.Props{"abs_1": "DOC1^Smith^John", "abs_2": "MED"}).
    BuildSegment("ADJ", hl7.Props{ /* … */ })
```

Field-level enforcement is identical to the typed methods: required fields throw if missing, withdrawn fields throw if set, deprecated (B) fields warn but still serialize. See the [Validation & errors section](../builder/index.md#-validation--errors) in the builder docs for the full per-code behavior.

---

## 📑 HL7 value tables (version-aware enforcement)

Many HL7 fields and composite components are bound to an **HL7 value table** — a fixed set of allowed codes (e.g. table `0001` Sex = `F`/`M`/`O`/`U`, table `0003` Event Type, table `0125` Value Type). go-hl7 ships the **complete** set of HL7-defined value tables, generated from the [Caristix HL7 Definition API](https://hl7-definition.caristix.com/v2/) and committed to the repo (no runtime network calls). They live in the `client/hl7/tables` package as `tables.TABLES`.

This goes **beyond** the upstream library, which only ever hand-populated a couple dozen tables. Every table-bound field (and every composite component) is now validated against its table automatically.

### Version-aware by design

HL7 value sets change between versions, so the registry is keyed **version → table id → ordered codes**:

```go
import "github.com/Bugs5382/go-hl7/client/hl7/tables"

tables.TABLES["2.1"]["0125"] // → [AD CK FT PN ST TM TS TX]  (v2.1 value types)
tables.TABLES["2.8"]["0125"] // → [AD CE CF … NM … XPN XTN]  (v2.8 value types)
```

The builder always validates against the table set for **its** version. A value that is valid in one version can be rejected in another:

```go
b28 := hl7.NewHL7_2_8()
b28.BuildMSH(/* … */)
b28.BuildOBX(hl7.Props{"obx_2": "NM", /* … */}) // ✅ NM is a v2.8 value type

b21 := hl7.NewHL7_2_1()
b21.BuildMSH(/* … */)
b21.BuildOBX(hl7.Props{"obx_2": "NM", /* … */}) // ❌ panics: Field 2 must be one of: AD, CK, FT, PN, ST, TM, TS, TX
```

### Hard error, no lenient mode

An out-of-table value is a **hard `HL7ValidationError`** — the same rejection path as the usage-code checks (there is no lenient mode in go-hl7). This applies to both field-level bindings and composite **component** bindings (for example, `PID.11` address `addressType` is validated against table `0190`).

### Tables with no fixed value set are not enforced

Some HL7 tables are user-defined or carry no published code list for a given version (e.g. `0296` Primary Language). When a table has **no values** for the active version, the field is **not** enforced — there is nothing to check against — so any value is accepted:

```go
b := hl7.NewHL7_2_1()
b.BuildMSH(/* … */)
b.BuildPID(hl7.Props{"pid_15": "anything-goes", /* … */}) // ✅ table 0296 is empty in v2.1
```

---

## 🧱 Always-available (base) segments

These four are implemented directly on `HL7_BASE` (not gated by version) and are available on every `NewHL7_2_x` builder.

| Segment | Builder | Purpose | Caristix |
|:---:|---|---|:---:|
| **ADD** | `BuildADD(props)` | Addendum — used to extend a message that exceeded the size limit. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/ADD) |
| **DSP** | `BuildDSP(props)` | Display Data — formatted display lines. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/DSP) |
| **NCK** | `BuildNCK()` | System Clock — synchronizes server clocks. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/NCK) |
| **NST** | `BuildNST(props)` | Application Control-level Statistics. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/NST) |

---

## 📊 Compatibility matrix

| Segment | 2.1 | 2.2 | 2.3 | 2.3.1 | 2.4 | 2.5 | 2.5.1 | 2.6 | 2.7 | 2.7.1 | 2.8 |
|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **ACC** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **AIG** | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **AIL** | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **AIP** | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **AIS** | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **AL1** | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **APR** | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **BLG** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **BPX** | ➖ | ➖ | ➖ | ➖ | ➖ | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ |
| **BTX** | ➖ | ➖ | ➖ | ➖ | ➖ | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ |
| **CSP** | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **CSR** | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **CSS** | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **CTD** | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **DG1** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **DRG** | ➖ | ➖ | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **DSC** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **ERR** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **EVN** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **FT1** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **GOL** | ➖ | ➖ | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **GT1** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **IAM** | ➖ | ➖ | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **IN1** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **IPC** | ➖ | ➖ | ➖ | ➖ | ➖ | ➖ | ➖ | ➖ | ✅ | ✅ | ✅ |
| **ISD** | ➖ | ➖ | ➖ | ➖ | ➖ | ➖ | ➖ | ➖ | ✅ | ✅ | ✅ |
| **ITM** | ➖ | ➖ | ➖ | ➖ | ➖ | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ |
| **IVT** | ➖ | ➖ | ➖ | ➖ | ➖ | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ |
| **MFE** | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **MFI** | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **MRG** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **MSA** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **MSH** | ✅ | 🔁 | 🔁 | 🔁 | 🔁 | 🔁 | 🔁 | 🔁 | 🔁 | 🔁 | 🔁 |
| **NK1** | ✅ | ✅ | 🔁 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **NPU** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **NSC** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **NTE** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **OBR** | ✅ | 🔁 | 🔁 | ✅ | 🔁 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **OBX** | ✅ | 🔁 | 🔁 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **ODS** | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **ODT** | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **OM1** | ➖ | ➖ | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **OM2** | ➖ | ➖ | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **OM3** | ➖ | ➖ | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **OM4** | ➖ | ➖ | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **OM5** | ➖ | ➖ | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **OM6** | ➖ | ➖ | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **ORC** | ✅ | 🔁 | 🔁 | ✅ | 🔁 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **PCR** | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **PD1** | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **PID** | ✅ | 🔁 | 🔁 | ✅ | 🔁 | 🔁 | ✅ | ✅ | ✅ | ✅ | ✅ |
| **PR1** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **PRA** | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **PRB** | ➖ | ➖ | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **PRD** | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **PSH** | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **PTH** | ➖ | ➖ | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **PV1** | ✅ | 🔁 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **QRD** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **QRF** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **RDF** | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **RDT** | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **REL** | ➖ | ➖ | ➖ | ➖ | ➖ | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ |
| **RGS** | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **ROL** | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **RX1** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **RXA** | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **RXD** | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **RXE** | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **RXG** | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **RXO** | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **RXR** | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **SCH** | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **SFT** | ➖ | ➖ | ➖ | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **SPM** | ➖ | ➖ | ➖ | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **STF** | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **STZ** | ➖ | ➖ | ➖ | ➖ | ➖ | ➖ | ➖ | ➖ | ➖ | ➖ | ✅ |
| **TXA** | ➖ | ➖ | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **UB1** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **UB2** | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **URD** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **URS** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **VAR** | ➖ | ➖ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

---

## 🗂️ By category

Group these alongside the workflow they support — easier than scrolling the matrix.

### 📨 Message control
**MSH**, **MSA**, **ERR**, **NCK**, **NST**, **NSC**, **DSC**, **DSP**, **ADD**, **SFT**

### 🧑 Patient demographics
**PID**, **PD1**, **NK1**, **GT1**, **IN1**, **MRG**, **AL1**, **IAM**

### 🏨 Patient visit / ADT
**EVN**, **PV1**, **DG1**, **DRG**, **PR1**, **NPU**, **OBX**, **NTE**

### 🔬 Orders & results (Lab / Rad)
**ORC**, **OBR**, **OBX**, **SPM**, **TXA**, **PRB**, **GOL**, **PTH**, **VAR**, **OM1**–**OM6**

### 💊 Pharmacy
**RXA**, **RXD**, **RXE**, **RXG**, **RXO**, **RXR**, **RX1** (HL7 2.1 only — replaced by RXO/RXE)

### 📅 Scheduling
**SCH**, **AIG**, **AIL**, **AIP**, **AIS**, **APR**, **RGS**

### 🧪 Clinical study
**CSP**, **CSR**, **CSS**

### 📁 Master files & staff
**MFE**, **MFI**, **STF**, **PRA**, **PRD**, **ROL**, **CTD**, **PSH**

### 🩸 Blood / inventory (HL7 2.6+)
**BPX**, **BTX**, **ITM**, **IVT**, **REL**

### 💰 Financial
**FT1**, **BLG**, **UB1**, **UB2**

### 🚑 Other / specialized
**ACC**, **PCR**, **QRD**, **QRF**, **RDF**, **RDT**, **URD**, **URS**, **ODS**, **ODT**, **IPC**, **ISD**, **STZ**

---

## 📋 Per-segment cheat-sheet

Each entry shows the **builder method**, **first-supported HL7 version**, a one-line description, and a link to Caristix for the full field reference. Where a segment is overridden by a later version (🔁 in the matrix) the library uses whichever version you instantiated.

| Segment | Builder | Since | Description | Caristix |
|:---:|---|:---:|---|:---:|
| ACC | `BuildACC(props)` | 2.1 | Accident details. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/ACC) |
| ADD | `BuildADD(props)` | base | Addendum continuation. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/ADD) |
| AIG | `BuildAIG(props)` | 2.3 | Appointment info — general resource. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/AIG) |
| AIL | `BuildAIL(props)` | 2.3 | Appointment info — location resource. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/AIL) |
| AIP | `BuildAIP(props)` | 2.3 | Appointment info — personnel resource. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/AIP) |
| AIS | `BuildAIS(props)` | 2.3 | Appointment info — service. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/AIS) |
| AL1 | `BuildAL1(props)` | 2.2 | Patient allergy information. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/AL1) |
| APR | `BuildAPR(props)` | 2.3 | Appointment preferences. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/APR) |
| BLG | `BuildBLG(props)` | 2.1 | Billing. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/BLG) |
| BPX | `BuildBPX(props)` | 2.6 | Blood product dispense status. | [link](https://hl7-definition.caristix.com/v2/HL7v2.6/Segments/BPX) |
| BTX | `BuildBTX(props)` | 2.6 | Blood product transfusion / disposition. | [link](https://hl7-definition.caristix.com/v2/HL7v2.6/Segments/BTX) |
| CSP | `BuildCSP(props)` | 2.3 | Clinical study phase. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/CSP) |
| CSR | `BuildCSR(props)` | 2.3 | Clinical study registration. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/CSR) |
| CSS | `BuildCSS(props)` | 2.3 | Clinical study data schedule. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/CSS) |
| CTD | `BuildCTD(props)` | 2.3 | Contact data. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/CTD) |
| DG1 | `BuildDG1(props)` | 2.1 | Diagnosis. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/DG1) |
| DRG | `BuildDRG(props)` | 2.4 | Diagnosis-related group. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/DRG) |
| DSC | `BuildDSC(props)` | 2.1 | Continuation pointer. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/DSC) |
| DSP | `BuildDSP(props)` | base | Display data. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/DSP) |
| ERR | `BuildERR(props)` | 2.1 | Error. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/ERR) |
| EVN | `BuildEVN(props)` | 2.1 | Event type (ADT trigger). | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/EVN) |
| FT1 | `BuildFT1(props)` | 2.1 | Financial transaction. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/FT1) |
| GOL | `BuildGOL(props)` | 2.4 | Goal detail. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/GOL) |
| GT1 | `BuildGT1(props)` | 2.1 | Guarantor. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/GT1) |
| IAM | `BuildIAM(props)` | 2.4 | Patient adverse reaction information. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/IAM) |
| IN1 | `BuildIN1(props)` | 2.1 | Insurance. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/IN1) |
| IPC | `BuildIPC(props)` | 2.7 | Imaging procedure control. | [link](https://hl7-definition.caristix.com/v2/HL7v2.7/Segments/IPC) |
| ISD | `BuildISD(props)` | 2.7 | Interaction status detail. | [link](https://hl7-definition.caristix.com/v2/HL7v2.7/Segments/ISD) |
| ITM | `BuildITM(props)` | 2.6 | Material item. | [link](https://hl7-definition.caristix.com/v2/HL7v2.6/Segments/ITM) |
| IVT | `BuildIVT(props)` | 2.6 | Material location. | [link](https://hl7-definition.caristix.com/v2/HL7v2.6/Segments/IVT) |
| MFE | `BuildMFE(props)` | 2.2 | Master file entry. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/MFE) |
| MFI | `BuildMFI(props)` | 2.2 | Master file identification. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/MFI) |
| MRG | `BuildMRG(props)` | 2.1 | Merge patient information. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/MRG) |
| MSA | `BuildMSA(props)` | 2.1 | Message acknowledgment. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/MSA) |
| MSH | `BuildMSH(props)` | 2.1 | Message header — **required first call**. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/MSH) |
| NCK | `BuildNCK()` | base | System clock sync. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/NCK) |
| NK1 | `BuildNK1(props)` | 2.1 | Next of kin / associated parties. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/NK1) |
| NPU | `BuildNPU(props)` | 2.1 | Bed status update. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/NPU) |
| NSC | `BuildNSC(props)` | 2.1 | Application status change. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/NSC) |
| NST | `BuildNST(props)` | base | Application control-level statistics. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/NST) |
| NTE | `BuildNTE(props)` | 2.1 | Notes & comments. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/NTE) |
| OBR | `BuildOBR(props)` | 2.1 | Observation request. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/OBR) |
| OBX | `BuildOBX(props)` | 2.1 | Observation / result. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/OBX) |
| ODS | `BuildODS(props)` | 2.2 | Dietary orders & supplements. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/ODS) |
| ODT | `BuildODT(props)` | 2.2 | Diet tray instructions. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/ODT) |
| OM1 | `BuildOM1(props)` | 2.4 | Master file — observation. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/OM1) |
| OM2 | `BuildOM2(props)` | 2.4 | Master file — numeric observation. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/OM2) |
| OM3 | `BuildOM3(props)` | 2.4 | Master file — categorical observation. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/OM3) |
| OM4 | `BuildOM4(props)` | 2.4 | Master file — observations that require specimens. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/OM4) |
| OM5 | `BuildOM5(props)` | 2.4 | Master file — observation batteries. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/OM5) |
| OM6 | `BuildOM6(props)` | 2.4 | Master file — observations as calculations. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/OM6) |
| ORC | `BuildORC(props)` | 2.1 | Common order. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/ORC) |
| PCR | `BuildPCR(props)` | 2.3 | Possible causal relationship. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/PCR) |
| PD1 | `BuildPD1(props)` | 2.3 | Patient additional demographic. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/PD1) |
| PID | `BuildPID(props)` | 2.1 | Patient identification. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/PID) |
| PR1 | `BuildPR1(props)` | 2.1 | Procedures. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/PR1) |
| PRA | `BuildPRA(props)` | 2.3 | Practitioner detail. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/PRA) |
| PRB | `BuildPRB(props)` | 2.4 | Problem detail. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/PRB) |
| PRD | `BuildPRD(props)` | 2.3 | Provider data. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/PRD) |
| PSH | `BuildPSH(props)` | 2.3 | Product summary header. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/PSH) |
| PTH | `BuildPTH(props)` | 2.4 | Pathway. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/PTH) |
| PV1 | `BuildPV1(props)` | 2.1 | Patient visit. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/PV1) |
| QRD | `BuildQRD(props)` | 2.1 | Query definition (original style). | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/QRD) |
| QRF | `BuildQRF(props)` | 2.1 | Query filter (original style). | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/QRF) |
| RDF | `BuildRDF(props)` | 2.3 | Table row definition. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/RDF) |
| RDT | `BuildRDT(props)` | 2.3 | Table row data. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/RDT) |
| REL | `BuildREL(props)` | 2.6 | Clinical relationship. | [link](https://hl7-definition.caristix.com/v2/HL7v2.6/Segments/REL) |
| RGS | `BuildRGS(props)` | 2.3 | Resource group. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/RGS) |
| ROL | `BuildROL(props)` | 2.3 | Role. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/ROL) |
| RX1 | `BuildRX1(props)` | 2.1 | Pharmacy order (legacy — replaced in 2.2 by RXO/RXE). | [link](https://hl7-definition.caristix.com/v2/HL7v2.1/Segments/RX1) |
| RXA | `BuildRXA(props)` | 2.2 | Pharmacy / treatment administration. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/RXA) |
| RXD | `BuildRXD(props)` | 2.2 | Pharmacy / treatment dispense. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/RXD) |
| RXE | `BuildRXE(props)` | 2.2 | Pharmacy / treatment encoded order. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/RXE) |
| RXG | `BuildRXG(props)` | 2.2 | Pharmacy / treatment give. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/RXG) |
| RXO | `BuildRXO(props)` | 2.2 | Pharmacy / treatment order. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/RXO) |
| RXR | `BuildRXR(props)` | 2.2 | Pharmacy / treatment route. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/RXR) |
| SCH | `BuildSCH(props)` | 2.3 | Scheduling activity information. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/SCH) |
| SFT | `BuildSFT(props)` | 2.5 | Software segment (sender identification). | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/SFT) |
| SPM | `BuildSPM(props)` | 2.5 | Specimen. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/SPM) |
| STF | `BuildSTF(props)` | 2.2 | Staff identification. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/STF) |
| STZ | `BuildSTZ(props)` | 2.8 | Sterilization parameter. | [link](https://hl7-definition.caristix.com/v2/HL7v2.8/Segments/STZ) |
| TXA | `BuildTXA(props)` | 2.4 | Transcription document header. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/TXA) |
| UB1 | `BuildUB1(props)` | 2.1 | UB-82 data. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/UB1) |
| UB2 | `BuildUB2(props)` | 2.2 | UB-92 data. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/UB2) |
| URD | `BuildURD(props)` | 2.1 | Results / update definition. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/URD) |
| URS | `BuildURS(props)` | 2.1 | Unsolicited selection. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/URS) |
| VAR | `BuildVAR(props)` | 2.3 | Variance. | [link](https://hl7-definition.caristix.com/v2/HL7v2.5/Segments/VAR) |

> 🧱 Need a Z-segment (vendor-specific)? The typed builders don't cover those by design. Use [`msg.AddSegment("ZXX")`](../builder/index.md#-direct-edits-with-msgset) on the `*builder.Message` returned from `ToMessage()`.
