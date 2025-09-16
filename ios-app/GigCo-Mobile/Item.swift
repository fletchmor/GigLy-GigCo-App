//
//  Item.swift
//  GigCo-Mobile
//
//  Created by Fletcher Morris on 9/8/25.
//

import Foundation
import SwiftData

@Model
final class Item {
    var timestamp: Date
    
    init(timestamp: Date) {
        self.timestamp = timestamp
    }
}
