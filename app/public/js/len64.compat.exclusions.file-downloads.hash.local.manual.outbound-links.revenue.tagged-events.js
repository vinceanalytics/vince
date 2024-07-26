!function(){"use strict";var u=window.location,o=window.document,l=o.getElementById("plausible"),p=l.getAttribute("data-api")||(i=(i=l).src.split("/"),n=i[0],i=i[2],n+"//"+i+"/api/event");function s(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){try{if("true"===window.localStorage.plausible_ignore)return s("localStorage flag",t)}catch(e){}var r=l&&l.getAttribute("data-include"),a=l&&l.getAttribute("data-exclude");if("pageview"===e){r=!r||r.split(",").some(n),a=a&&a.split(",").some(n);if(!r||a)return s("exclusion rule",t)}function n(e){var t=u.pathname;return(t+=u.hash).match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var r={},i=(r.n=e,r.u=t&&t.u?t.u:u.href,r.d=l.getAttribute("data-domain"),r.r=o.referrer||null,t&&t.meta&&(r.m=JSON.stringify(t.meta)),t&&t.props&&(r.p=t.props),t&&t.revenue&&(r.$=t.revenue),r.h=1,new XMLHttpRequest);i.open("POST",p,!0),i.setRequestHeader("Content-Type","text/plain"),i.send(JSON.stringify(r)),i.onreadystatechange=function(){4===i.readyState&&t&&t.callback&&t.callback({status:i.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var r=0;r<t.length;r++)e.apply(this,t[r]);function c(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}var f=1;function a(e){if("auxclick"!==e.type||e.button===f){var t,r,a=function(e){for(;e&&(void 0===e.tagName||!c(e)||!e.href);)e=e.parentNode;return e}(e.target),n=a&&a.href&&a.href.split("?")[0];if(!function e(t,r){if(!t||b<r)return!1;if(w(t))return!0;return e(t.parentNode,r+1)}(a,0))return(t=a)&&t.href&&t.host&&t.host!==u.host?v(e,a,{name:"Outbound Link: Click",props:{url:a.href}}):(t=n)&&(r=t.split(".").pop(),d.some(function(e){return e===r}))?v(e,a,{name:"File Download",props:{url:n}}):void 0}}function v(e,t,r){var a,n=!1;function i(){n||(n=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?((a={props:r.props}).revenue=r.revenue,plausible(r.name,a)):((a={props:r.props,callback:i}).revenue=r.revenue,plausible(r.name,a),setTimeout(i,5e3),e.preventDefault())}o.addEventListener("click",a),o.addEventListener("auxclick",a);var n=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],i=l.getAttribute("file-types"),m=l.getAttribute("add-file-types"),d=i&&i.split(",")||m&&m.split(",").concat(n)||n;function g(e){var e=w(e)?e:e&&e.parentNode,t={name:null,props:{},revenue:{}},r=e&&e.classList;if(r)for(var a=0;a<r.length;a++){var n,i,u=r.item(a),o=u.match(/plausible-event-(.+)(=|--)(.+)/),o=(o&&(n=o[1],i=o[3].replace(/\+/g," "),"name"==n.toLowerCase()?t.name=i:t.props[n]=i),u.match(/plausible-revenue-(.+)(=|--)(.+)/));o&&(n=o[1],i=o[3],t.revenue[n]=i)}return t}var b=3;function h(e){if("auxclick"!==e.type||e.button===f){for(var t,r,a,n,i=e.target,u=0;u<=b&&i;u++){if((a=i)&&a.tagName&&"form"===a.tagName.toLowerCase())return;c(i)&&(t=i),w(i)&&(r=i),i=i.parentNode}r&&(n=g(r),t?(n.props.url=t.href,v(e,t,n)):((e={}).props=n.props,e.revenue=n.revenue,plausible(n.name,e)))}}function w(e){var t=e&&e.classList;if(t)for(var r=0;r<t.length;r++)if(t.item(r).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}o.addEventListener("submit",function(e){var t,r=e.target,a=g(r);function n(){t||(t=!0,r.submit())}a.name&&(e.preventDefault(),t=!1,setTimeout(n,5e3),(e={props:a.props,callback:n}).revenue=a.revenue,plausible(a.name,e))}),o.addEventListener("click",h),o.addEventListener("auxclick",h)}();