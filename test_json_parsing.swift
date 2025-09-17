import Foundation

struct JobCreateResponse: Codable {
    let id: Int
    let uuid: String
    let title: String
    let description: String
    let category: String?
    let locationAddress: String?
    let totalPay: Double?
    let status: String
    let consumerID: Int?
    let scheduledStart: String?
    let createdAt: String
    let updatedAt: String

    enum CodingKeys: String, CodingKey {
        case id, uuid, title, description, category, status
        case locationAddress = "location_address"
        case totalPay = "total_pay"
        case consumerID = "consumer_id"
        case scheduledStart = "scheduled_start"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }
}

let jsonString = """
{"id":6,"uuid":"bf7ce14a-3d12-4887-9e47-7bf0b8f52a4e","consumer_id":2,"title":"Cleaning","description":"Clean my house now!","category":"Cleaning","location_address":"Test Address","total_pay":67,"status":"posted","scheduled_start":"2025-09-17T00:41:00Z","created_at":"2025-09-17T00:31:08.235767Z","updated_at":"2025-09-17T00:31:08.235767Z"}
"""

if let data = jsonString.data(using: .utf8) {
    do {
        let response = try JSONDecoder().decode(JobCreateResponse.self, from: data)
        print("✅ Parsing successful!")
        print("ID: \(response.id)")
        print("Consumer ID: \(response.consumerID ?? -1)")
        print("Title: \(response.title)")
    } catch {
        print("❌ Parsing failed: \(error)")
    }
} else {
    print("❌ Failed to create data from string")
}
