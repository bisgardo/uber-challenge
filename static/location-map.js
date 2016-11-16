'use strict';

function initMap() {
	var sf = {lat: 37.749, lng: -122.439};
	var mapElement = document.getElementById('map');
	var map = new google.maps.Map(mapElement, {
		zoom: 11,
		center: sf
	});
	
	var $locations = $('.location');
	
	var bounds = new google.maps.LatLngBounds();
	$locations.each(function () {
		var $location = $(this);
		var name = $location.data('name');
		var lat = $location.data('lat');
		var lng = $location.data('lng');
		
		if (lat && lng) {
			var marker = new google.maps.Marker({
				title: name,
				position: {lat: lat, lng: lng},
				// animation: google.maps.Animation.DROP,
				map: map
			});
			
			$location.hover(function () {
				// TODO Zoom out until marker is visible.
				marker.setAnimation(google.maps.Animation.BOUNCE);
				$location.addClass('secondary');
			}, function () {
				marker.setAnimation(null);
				$location.removeClass('secondary');
			});
			
			marker.addListener('mouseover', function () {
				$location.addClass('primary');
				var locationElement = $location[0];
				locationElement.scrollIntoView && locationElement.scrollIntoView({
					behavior: 'smooth'
				});
			});
			marker.addListener('mouseout', function () {
				$location.removeClass('primary');
			});
			
			bounds.extend(marker.getPosition());
		} else {
			$location.addClass('warning').attr('title', 'Could not find coordinates for this location...');
		}
	});
	
	if (!bounds.isEmpty()) {
		bounds.extend(sf);
		map.fitBounds(bounds);
	}
}
